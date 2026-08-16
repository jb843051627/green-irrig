package model

import "time"

// Zone 灌溉区域
type Zone struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	CropType  string    `gorm:"size:50" json:"crop_type"`
	AreaM2    float64   `json:"area_m2"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// IrrigationPlan 灌溉计划
type IrrigationPlan struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ZoneID       uint      `gorm:"not null" json:"zone_id"`
	VolumeLiters float64   `json:"volume_liters"`
	ScheduledAt  time.Time `json:"scheduled_at"`
	Status       string    `gorm:"size:20;default:'pending'" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// SensorReading 传感器读数
type SensorReading struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ZoneID    uint      `gorm:"not null" json:"zone_id"`
	SoilMoist float64   `json:"soil_moist"`
	TempC     float64   `json:"temp_c"`
	ReadAt    time.Time `json:"read_at"`
}
