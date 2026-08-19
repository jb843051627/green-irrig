package service

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"green-irrig/internal/model"
)

// newTestService 创建独立临时文件库的测试服务（避免 -count=N 迭代间共享内存库残留数据）
func newTestService(t *testing.T) *IrrigationService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	// Windows 上文件句柄不关会导致 TempDir 清理失败
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	})
	if err := db.AutoMigrate(&model.Zone{}, &model.IrrigationPlan{}, &model.SensorReading{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewIrrigationService(db)
}

// ScheduleIrrigation 对不存在区域返回的错误应保留 ErrZoneNotFound 错误链
func TestBug04_ScheduleIrrigationErrorChainPreserved(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.ScheduleIrrigation(context.Background(), 9999, 10.0, time.Now().Add(1*time.Hour))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrZoneNotFound) {
		t.Fatalf("want errors.Is(err, ErrZoneNotFound)=true, got false; err=%v", err)
	}
}