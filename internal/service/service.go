package service

import (
	"context"
	"errors"
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
}

func NewIrrigationService(db *gorm.DB) *IrrigationService {
	return &IrrigationService{db: db}
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
		return nil, err
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
	return reading, nil
}

// ExecuteIrrigation 执行灌溉
func (s *IrrigationService) ExecuteIrrigation(ctx context.Context, planID uint) error {
	var plan model.IrrigationPlan
	if err := s.db.WithContext(ctx).First(&plan, planID).Error; err != nil {
		return err
	}
	plan.Status = "completed"
	return s.db.WithContext(ctx).Save(&plan).Error
}
