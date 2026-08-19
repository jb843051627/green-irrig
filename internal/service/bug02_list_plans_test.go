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

// ListPlans 返回的切片不应与内部缓存共享底层数组：
// 调用方修改返回切片的元素后，再次 ListPlans 不应看到被污染的数据
func TestBug02_ListPlansNoCachePollution(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	// 造一个 zone 和一个 plan
	zone := &model.Zone{Name: "A区", CropType: "番茄", AreaM2: 100, Active: true}
	if err := svc.db.Create(zone).Error; err != nil {
		t.Fatalf("create zone: %v", err)
	}
	plan := &model.IrrigationPlan{ZoneID: zone.ID, VolumeLiters: 10, ScheduledAt: time.Now(), Status: "pending"}
	if err := svc.db.Create(plan).Error; err != nil {
		t.Fatalf("create plan: %v", err)
	}

	plans1, err := svc.ListPlans(ctx, zone.ID)
	if err != nil {
		t.Fatalf("list plans: %v", err)
	}
	if len(plans1) != 1 {
		t.Fatalf("want 1 plan, got %d", len(plans1))
	}

	// 调用方修改返回切片的元素（模拟下游改动）
	plans1[0].VolumeLiters = 999

	// 再次查询：不应看到被污染的数据
	plans2, err := svc.ListPlans(ctx, zone.ID)
	if err != nil {
		t.Fatalf("list plans again: %v", err)
	}
	if len(plans2) != 1 {
		t.Fatalf("want 1 plan, got %d", len(plans2))
	}
	if plans2[0].VolumeLiters == 999 {
		t.Fatalf("cache polluted: plans2[0].VolumeLiters=%v, want original 10", plans2[0].VolumeLiters)
	}
}