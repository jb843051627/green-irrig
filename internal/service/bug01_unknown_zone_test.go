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

// 瀵逛笉瀛樺湪鐨勫尯鍩熻皟搴︾亴婧夛細涓嶅簲 panic锛屽簲杩斿洖 ErrZoneNotFound
func TestBug01_ScheduleIrrigationUnknownZoneNoPanic(t *testing.T) {
	svc := newTestService(t)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	_, err := svc.ScheduleIrrigation(context.Background(), 9999, 10.0, time.Now().Add(1*time.Hour))
	if !errors.Is(err, ErrZoneNotFound) {
		t.Fatalf("want ErrZoneNotFound, got %v", err)
	}
}
