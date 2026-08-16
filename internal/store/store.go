package store

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"green-irrig/internal/model"
)

// OpenDB 打开 SQLite 数据库（内存模式）
func OpenDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&model.Zone{}, &model.IrrigationPlan{}, &model.SensorReading{}); err != nil {
		return nil, err
	}
	return db, nil
}
