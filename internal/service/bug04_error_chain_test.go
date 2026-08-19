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

// ScheduleIrrigation 瀵逛笉瀛樺湪鍖哄煙杩斿洖鐨勯敊璇簲淇濈暀 ErrZoneNotFound 閿欒閾?func TestBug04_ScheduleIrrigationErrorChainPreserved(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.ScheduleIrrigation(context.Background(), 9999, 10.0, time.Now().Add(1*time.Hour))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrZoneNotFound) {
		t.Fatalf("want errors.Is(err, ErrZoneNotFound)=true, got false; err=%v", err)
	}
}
