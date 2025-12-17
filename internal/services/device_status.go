package services

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// DeviceStatusService 设备状态服务
type DeviceStatusService struct {
	db          *gorm.DB
	stopChan    chan struct{}
	offlineTime time.Duration // 离线判定时间
}

// NewDeviceStatusService 创建设备状态服务
func NewDeviceStatusService(db *gorm.DB) *DeviceStatusService {
	return &DeviceStatusService{
		db:          db,
		stopChan:    make(chan struct{}),
		offlineTime: 3 * time.Minute, // 3分钟无活动视为离线
	}
}

// Start 启动设备状态监控服务
func (s *DeviceStatusService) Start() {
	log.Println("设备状态监控服务已启动，离线判定时间: 3分钟")

	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkDeviceStatus()
		case <-s.stopChan:
			log.Println("设备状态监控服务已停止")
			return
		}
	}
}

// Stop 停止服务
func (s *DeviceStatusService) Stop() {
	close(s.stopChan)
}

// checkDeviceStatus 检查并更新设备状态
func (s *DeviceStatusService) checkDeviceStatus() {
	offlineThreshold := time.Now().Add(-s.offlineTime)

	// 将超过3分钟未活动的非离线设备标记为离线
	result := s.db.Exec(`
		UPDATE devices 
		SET status = 'offline' 
		WHERE status != 'offline' 
		  AND last_heartbeat < ?
	`, offlineThreshold)

	if result.Error != nil {
		log.Printf("更新设备离线状态失败: %v", result.Error)
		return
	}

	if result.RowsAffected > 0 {
		log.Printf("已将 %d 台设备标记为离线（超过3分钟无活动）", result.RowsAffected)
	}
}
