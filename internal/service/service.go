package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
	"green-irrig/internal/model"
)

var (
	ErrZoneNotFound  = errors.New("zone not found")
	ErrInvalidVolume = errors.New("invalid volume")
)

// IrrigationService 灌溉服务
type IrrigationService struct {
	db *gorm.DB
	// 区域最新读数缓存（zoneID -> reading）
	latestReadings map[uint]*model.SensorReading
	mu             sync.RWMutex
}

func NewIrrigationService(db *gorm.DB) *IrrigationService {
	return &IrrigationService{db: db, latestReadings: make(map[uint]*model.SensorReading)}
}

// GetZoneByID 按 ID 查询区域
func (s *IrrigationService) GetZoneByID(ctx context.Context, id uint) (*model.Zone, error) {
	var zone model.Zone
	result := s.db.WithContext(ctx).First(&zone, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrZoneNotFound
		}
		return nil, result.Error
	}
	return &zone, nil
}

// ScheduleIrrigation 创建灌溉计划
func (s *IrrigationService) ScheduleIrrigation(ctx context.Context, zoneID uint, volume float64, when time.Time) (*model.IrrigationPlan, error) {
	zone, err := s.GetZoneByID(ctx, zoneID)
	if err != nil {
		return nil, fmt.Errorf("schedule irrigation: %w", err)
	}
	if !zone.Active {
		return nil, ErrZoneNotFound
	}
	if volume <= 0 {
		return nil, ErrInvalidVolume
	}
	plan := &model.IrrigationPlan{
		ZoneID:       zoneID,
		VolumeLiters: volume,
		ScheduledAt:  when,
		Status:       "pending",
	}
	if err := s.db.WithContext(ctx).Create(plan).Error; err != nil {
		return nil, err
	}
	return plan, nil
}

// ListPlans 列出灌溉计划
func (s *IrrigationService) ListPlans(ctx context.Context, zoneID uint) ([]model.IrrigationPlan, error) {
	var plans []model.IrrigationPlan
	if err := s.db.WithContext(ctx).Where("zone_id = ?", zoneID).Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

// RecordReading 记录传感器读数
func (s *IrrigationService) RecordReading(ctx context.Context, zoneID uint, moist, temp float64) (*model.SensorReading, error) {
	reading := &model.SensorReading{
		ZoneID:    zoneID,
		SoilMoist: moist,
		TempC:     temp,
		ReadAt:    time.Now(),
	}
	if err := s.db.WithContext(ctx).Create(reading).Error; err != nil {
		return nil, err
	}
	s.latestReadings[zoneID] = reading
	return reading, nil
}

// GetLatestReading 获取区域最新读数
func (s *IrrigationService) GetLatestReading(zoneID uint) *model.SensorReading {
	return s.latestReadings[zoneID]
}

// ExecuteIrrigation 执行灌溉（模拟 5 个灌溉周期，周期之间检查 ctx 是否取消）
func (s *IrrigationService) ExecuteIrrigation(ctx context.Context, planID uint) error {
	var plan model.IrrigationPlan
	if err := s.db.WithContext(ctx).First(&plan, planID).Error; err != nil {
		return err
	}
	const cycles = 5
	for i := 0; i < cycles; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		time.Sleep(20 * time.Millisecond) // 模拟单个周期的灌溉动作
	}
	plan.Status = "completed"
	return s.db.Save(&plan).Error
}
