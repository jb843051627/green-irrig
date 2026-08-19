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

// newTestService 鍒涘缓鐙珛涓存椂鏂囦欢搴撶殑娴嬭瘯鏈嶅姟锛堥伩鍏?-count=N 杩唬闂村叡浜唴瀛樺簱娈嬬暀鏁版嵁锛?func newTestService(t *testing.T) *IrrigationService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	// Windows 涓婃枃浠跺彞鏌勪笉鍏充細瀵艰嚧 TempDir 娓呯悊澶辫触
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

// ListPlans 杩斿洖鐨勫垏鐗囦笉搴斾笌鍐呴儴缂撳瓨鍏变韩搴曞眰鏁扮粍锛?// 璋冪敤鏂逛慨鏀硅繑鍥炲垏鐗囩殑鍏冪礌鍚庯紝鍐嶆 ListPlans 涓嶅簲鐪嬪埌琚薄鏌撶殑鏁版嵁
func TestBug02_ListPlansNoCachePollution(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	// 閫犱竴涓?zone 鍜屼竴涓?plan
	zone := &model.Zone{Name: "A鍖?, CropType: "鐣寗", AreaM2: 100, Active: true}
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

	// 璋冪敤鏂逛慨鏀硅繑鍥炲垏鐗囩殑鍏冪礌锛堟ā鎷熶笅娓告敼鍔級
	plans1[0].VolumeLiters = 999

	// 鍐嶆鏌ヨ锛氫笉搴旂湅鍒拌姹℃煋鐨勬暟鎹?	plans2, err := svc.ListPlans(ctx, zone.ID)
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
