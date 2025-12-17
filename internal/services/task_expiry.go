package services

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"jd-task-platform-go/internal/models"
)

// TaskExpiryService 任务过期检查服务
type TaskExpiryService struct {
	db       *gorm.DB
	interval time.Duration
	stopChan chan struct{}
}

// NewTaskExpiryService 创建任务过期检查服务
func NewTaskExpiryService(db *gorm.DB) *TaskExpiryService {
	return &TaskExpiryService{
		db:       db,
		interval: time.Minute, // 每分钟检查一次
		stopChan: make(chan struct{}),
	}
}

// Start 启动过期检查服务
func (s *TaskExpiryService) Start() {
	log.Println("✓ 任务过期检查服务已启动（每分钟检查一次）")
	go s.run()
}

// Stop 停止过期检查服务
func (s *TaskExpiryService) Stop() {
	close(s.stopChan)
	log.Println("任务过期检查服务已停止")
}

// run 运行过期检查循环
func (s *TaskExpiryService) run() {
	// 启动时立即执行一次检查
	s.checkExpiredTasks()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkExpiredTasks()
		case <-s.stopChan:
			return
		}
	}
}

// checkExpiredTasks 检查并处理过期任务
func (s *TaskExpiryService) checkExpiredTasks() {
	// 计算24小时前的时间点
	expireThreshold := time.Now().Add(-24 * time.Hour)

	// 查找过期任务：
	// 1. start_time + 24小时 < 当前时间
	// 2. 状态为 waiting 或 running
	// 3. executed_count < execute_count (未完成)
	var expiredTasks []models.Task
	if err := s.db.Where(
		"start_time < ? AND status IN (?, ?) AND executed_count < execute_count",
		expireThreshold, "waiting", "running",
	).Find(&expiredTasks).Error; err != nil {
		log.Printf("查询过期任务失败: %v", err)
		return
	}

	if len(expiredTasks) == 0 {
		return // 没有过期任务
	}

	log.Printf("发现 %d 个过期任务，开始处理退款...", len(expiredTasks))

	for _, task := range expiredTasks {
		s.processExpiredTask(&task)
	}
}

// processExpiredTask 处理单个过期任务
func (s *TaskExpiryService) processExpiredTask(task *models.Task) {
	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("处理过期任务 %d 时发生panic: %v", task.ID, r)
		}
	}()

	// 获取任务类型价格
	var taskType models.TaskType
	if err := tx.Where("type_code = ?", task.TaskType).First(&taskType).Error; err != nil {
		tx.Rollback()
		log.Printf("获取任务类型失败 (task_id=%d, type=%s): %v", task.ID, task.TaskType, err)
		return
	}

	// 计算退款金额
	// 未完成次数 = 总次数 - 已完成次数
	remainingCount := task.ExecuteCount - task.ExecutedCount
	refundAmount := remainingCount * taskType.JingdouPrice

	// 如果是管理员创建的任务(consume_jingdou=0)，不需要退款
	if task.ConsumeJingdou == 0 {
		refundAmount = 0
	} else {
		// 根据实际消耗比例计算退款
		// 实际退款 = 消耗京豆 * (未完成次数 / 总次数)
		if task.ExecuteCount > 0 {
			refundAmount = task.ConsumeJingdou * remainingCount / task.ExecuteCount
		}
	}

	// 获取用户
	var user models.User
	if err := tx.First(&user, task.UserID).Error; err != nil {
		tx.Rollback()
		log.Printf("获取用户失败 (task_id=%d, user_id=%d): %v", task.ID, task.UserID, err)
		return
	}

	// 更新用户余额
	if refundAmount > 0 {
		user.JingdouBalance += refundAmount
		if err := tx.Save(&user).Error; err != nil {
			tx.Rollback()
			log.Printf("更新用户余额失败 (task_id=%d, user_id=%d): %v", task.ID, task.UserID, err)
			return
		}

		// 创建京豆日志
		jingdouLog := models.JingdouLog{
			UserID:        user.ID,
			Amount:        refundAmount,
			Balance:       user.JingdouBalance,
			OperationType: "refund",
			RelatedID:     &task.ID,
			Remark:        "任务过期自动退款 - SKU:" + task.SKU + " (完成" + formatInt(task.ExecutedCount) + "/" + formatInt(task.ExecuteCount) + ")",
			CreatedAt:     time.Now(),
		}
		if err := tx.Create(&jingdouLog).Error; err != nil {
			tx.Rollback()
			log.Printf("创建京豆日志失败 (task_id=%d): %v", task.ID, err)
			return
		}
	}

	// 更新任务状态为 partial_completed
	task.Status = "partial_completed"
	task.UpdatedAt = time.Now()

	// 更新备注，记录过期处理信息
	expireRemark := "【系统自动处理】任务过期，完成" + formatInt(task.ExecutedCount) + "/" + formatInt(task.ExecuteCount) + "次"
	if refundAmount > 0 {
		expireRemark += "，退还" + formatInt(refundAmount) + "京豆"
	}
	if task.Remark != "" {
		task.Remark = task.Remark + " | " + expireRemark
	} else {
		task.Remark = expireRemark
	}

	if err := tx.Save(task).Error; err != nil {
		tx.Rollback()
		log.Printf("更新任务状态失败 (task_id=%d): %v", task.ID, err)
		return
	}

	// 创建任务日志
	taskLog := models.TaskLog{
		TaskID:    task.ID,
		Status:    "partial_completed",
		Message:   expireRemark,
		CreatedAt: time.Now(),
	}
	if err := tx.Create(&taskLog).Error; err != nil {
		tx.Rollback()
		log.Printf("创建任务日志失败 (task_id=%d): %v", task.ID, err)
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Printf("提交事务失败 (task_id=%d): %v", task.ID, err)
		return
	}

	log.Printf("任务过期处理完成: task_id=%d, sku=%s, 完成=%d/%d, 退款=%d京豆",
		task.ID, task.SKU, task.ExecutedCount, task.ExecuteCount, refundAmount)
}

// formatInt 格式化整数为字符串
func formatInt(n int) string {
	return fmt.Sprintf("%d", n)
}
