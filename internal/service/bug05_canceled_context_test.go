package service

import (
	"context"
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

// context 超时后执行灌溉：应中止（返回错误），不能把计划标记为 completed
func TestBug05_ExecuteIrrigationCanceledContext(t *testing.T) {
	svc := newTestService(t)

	plan := &model.IrrigationPlan{ZoneID: 1, VolumeLiters: 10, ScheduledAt: time.Now(), Status: "pending"}
	if err := svc.db.Create(plan).Error; err != nil {
		t.Fatalf("create plan: %v", err)
	}

	// 30ms 超时 < 5 周期 × 20ms = 100ms，循环中途会取消
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	err := svc.ExecuteIrrigation(ctx, plan.ID)
	if err == nil {
		t.Fatal("expected error for timed-out context, got nil")
	}

	// 验证计划未被标记完成（取消应中止执行）
	var got model.IrrigationPlan
	if err := svc.db.First(&got, plan.ID).Error; err != nil {
		t.Fatalf("reload plan: %v", err)
	}
	if got.Status == "completed" {
		t.Fatalf("plan marked completed despite timed-out context: %+v", got)
	}
}