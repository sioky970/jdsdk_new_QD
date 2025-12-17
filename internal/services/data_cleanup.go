package services

import (
	"log"
	"time"

	"gorm.io/gorm"

	"jd-task-platform-go/internal/models"
)

// DataCleanupService 数据清理服务
type DataCleanupService struct {
	db            *gorm.DB
	retentionDays int // 数据保留天数
	cleanupHour   int // 每天清理时间（小时）
	stopChan      chan struct{}
}

// NewDataCleanupService 创建数据清理服务
// retentionDays: 数据保留天数，默认60天
// cleanupHour: 每天清理时间，默认0点
func NewDataCleanupService(db *gorm.DB, retentionDays int, cleanupHour int) *DataCleanupService {
	if retentionDays <= 0 {
		retentionDays = 60
	}
	if cleanupHour < 0 || cleanupHour > 23 {
		cleanupHour = 0
	}
	return &DataCleanupService{
		db:            db,
		retentionDays: retentionDays,
		cleanupHour:   cleanupHour,
		stopChan:      make(chan struct{}),
	}
}

// Start 启动数据清理服务
func (s *DataCleanupService) Start() {
	log.Printf("✓ 数据清理服务已启动（保留%d天，每日%d:00执行）", s.retentionDays, s.cleanupHour)
	go s.run()
}

// Stop 停止数据清理服务
func (s *DataCleanupService) Stop() {
	close(s.stopChan)
	log.Println("数据清理服务已停止")
}

// run 运行清理调度
func (s *DataCleanupService) run() {
	for {
		// 计算下次清理时间
		now := time.Now()
		nextCleanup := time.Date(now.Year(), now.Month(), now.Day(), s.cleanupHour, 0, 0, 0, now.Location())

		// 如果今天的清理时间已过，则安排到明天
		if now.After(nextCleanup) {
			nextCleanup = nextCleanup.Add(24 * time.Hour)
		}

		waitDuration := nextCleanup.Sub(now)
		log.Printf("下次数据清理时间: %s (等待 %v)", nextCleanup.Format("2006-01-02 15:04:05"), waitDuration.Round(time.Minute))

		select {
		case <-time.After(waitDuration):
			s.executeCleanup()
		case <-s.stopChan:
			return
		}
	}
}

// executeCleanup 执行清理任务
func (s *DataCleanupService) executeCleanup() {
	startTime := time.Now()
	log.Println("========================================")
	log.Println("  开始执行数据清理任务")
	log.Println("========================================")

	// 计算清理阈值时间
	threshold := time.Now().AddDate(0, 0, -s.retentionDays)
	log.Printf("清理 %s 之前的数据（保留%d天）", threshold.Format("2006-01-02"), s.retentionDays)

	// 统计清理结果
	result := CleanupResult{}

	// 1. 清理任务日志（先清理，因为有外键关联）
	result.TaskLogsDeleted = s.cleanupTaskLogs(threshold)

	// 2. 清理任务（只清理已完成/已取消/部分完成的）
	result.TasksDeleted = s.cleanupTasks(threshold)

	// 3. 清理设备任务历史
	result.DeviceHistoryDeleted = s.cleanupDeviceTaskHistory(threshold)

	// 4. 清理API日志
	result.APILogsDeleted = s.cleanupAPILogs(threshold)

	duration := time.Since(startTime)

	log.Println("========================================")
	log.Printf("  数据清理完成，耗时: %v", duration.Round(time.Millisecond))
	log.Printf("  - 任务记录: %d 条", result.TasksDeleted)
	log.Printf("  - 任务日志: %d 条", result.TaskLogsDeleted)
	log.Printf("  - 设备历史: %d 条", result.DeviceHistoryDeleted)
	log.Printf("  - API日志: %d 条", result.APILogsDeleted)
	log.Println("========================================")
}

// CleanupResult 清理结果统计
type CleanupResult struct {
	TasksDeleted         int64
	TaskLogsDeleted      int64
	DeviceHistoryDeleted int64
	APILogsDeleted       int64
}

// cleanupTasks 清理过期任务
func (s *DataCleanupService) cleanupTasks(threshold time.Time) int64 {
	// 只清理已完成/已取消/部分完成/失败的任务
	// 不清理 waiting 和 running 状态的任务
	safeStatuses := []string{"completed", "cancelled", "partial_completed", "failed"}

	result := s.db.Where(
		"created_at < ? AND status IN ?",
		threshold, safeStatuses,
	).Delete(&models.Task{})

	if result.Error != nil {
		log.Printf("清理任务失败: %v", result.Error)
		return 0
	}

	return result.RowsAffected
}

// cleanupTaskLogs 清理过期任务日志
func (s *DataCleanupService) cleanupTaskLogs(threshold time.Time) int64 {
	// 清理关联的过期任务的日志
	// 使用子查询找出要删除的任务ID
	safeStatuses := []string{"completed", "cancelled", "partial_completed", "failed"}

	// 先查询要删除的任务ID
	var taskIDs []uint
	s.db.Model(&models.Task{}).
		Where("created_at < ? AND status IN ?", threshold, safeStatuses).
		Pluck("id", &taskIDs)

	if len(taskIDs) == 0 {
		return 0
	}

	// 删除这些任务的日志
	result := s.db.Where("task_id IN ?", taskIDs).Delete(&models.TaskLog{})

	if result.Error != nil {
		log.Printf("清理任务日志失败: %v", result.Error)
		return 0
	}

	return result.RowsAffected
}

// cleanupDeviceTaskHistory 清理设备任务历史
func (s *DataCleanupService) cleanupDeviceTaskHistory(threshold time.Time) int64 {
	result := s.db.Where("execute_time < ?", threshold).Delete(&models.DeviceTaskHistory{})

	if result.Error != nil {
		log.Printf("清理设备任务历史失败: %v", result.Error)
		return 0
	}

	return result.RowsAffected
}

// cleanupAPILogs 清理API调用日志
func (s *DataCleanupService) cleanupAPILogs(threshold time.Time) int64 {
	result := s.db.Where("created_at < ?", threshold).Delete(&models.APILog{})

	if result.Error != nil {
		log.Printf("清理API日志失败: %v", result.Error)
		return 0
	}

	return result.RowsAffected
}

// ManualCleanup 手动触发清理（供API调用）
func (s *DataCleanupService) ManualCleanup() CleanupResult {
	startTime := time.Now()
	log.Println("手动触发数据清理...")

	threshold := time.Now().AddDate(0, 0, -s.retentionDays)

	result := CleanupResult{}
	result.TaskLogsDeleted = s.cleanupTaskLogs(threshold)
	result.TasksDeleted = s.cleanupTasks(threshold)
	result.DeviceHistoryDeleted = s.cleanupDeviceTaskHistory(threshold)
	result.APILogsDeleted = s.cleanupAPILogs(threshold)

	log.Printf("手动清理完成，耗时: %v", time.Since(startTime).Round(time.Millisecond))

	return result
}

// GetRetentionDays 获取保留天数
func (s *DataCleanupService) GetRetentionDays() int {
	return s.retentionDays
}
