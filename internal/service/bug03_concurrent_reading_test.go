package service

import (
	"context"
	"path/filepath"
	"sync"
	"testing"

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

// 并发记录传感器读数：不应触发 data race
func TestBug03_RecordReadingConcurrentSafe(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	// 预造 10 个 zone
	for i := 0; i < 10; i++ {
		z := &model.Zone{Name: "Z", CropType: "番茄", AreaM2: 100, Active: true}
		if err := svc.db.Create(z).Error; err != nil {
			t.Fatalf("create zone: %v", err)
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, _ = svc.RecordReading(ctx, uint(n%10+1), 30.0+float64(n), 25.0)
		}(i)
	}
	wg.Wait()
}