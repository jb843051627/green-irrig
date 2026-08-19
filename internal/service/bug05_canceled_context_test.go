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

// context 瓒呮椂鍚庢墽琛岀亴婧夛細搴斾腑姝紙杩斿洖閿欒锛夛紝涓嶈兘鎶婅鍒掓爣璁颁负 completed
func TestBug05_ExecuteIrrigationCanceledContext(t *testing.T) {
	svc := newTestService(t)

	plan := &model.IrrigationPlan{ZoneID: 1, VolumeLiters: 10, ScheduledAt: time.Now(), Status: "pending"}
	if err := svc.db.Create(plan).Error; err != nil {
		t.Fatalf("create plan: %v", err)
	}

	// 30ms 瓒呮椂 < 5 鍛ㄦ湡 脳 20ms = 100ms锛屽惊鐜腑閫斾細鍙栨秷
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	err := svc.ExecuteIrrigation(ctx, plan.ID)
	if err == nil {
		t.Fatal("expected error for timed-out context, got nil")
	}

	// 楠岃瘉璁″垝鏈鏍囪瀹屾垚锛堝彇娑堝簲涓鎵ц锛?	var got model.IrrigationPlan
	if err := svc.db.First(&got, plan.ID).Error; err != nil {
		t.Fatalf("reload plan: %v", err)
	}
	if got.Status == "completed" {
		t.Fatalf("plan marked completed despite timed-out context: %+v", got)
	}
}
