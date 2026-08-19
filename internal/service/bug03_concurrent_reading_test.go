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

// 骞跺彂璁板綍浼犳劅鍣ㄨ鏁帮細涓嶅簲瑙﹀彂 data race
func TestBug03_RecordReadingConcurrentSafe(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	// 棰勯€?10 涓?zone
	for i := 0; i < 10; i++ {
		z := &model.Zone{Name: "Z", CropType: "鐣寗", AreaM2: 100, Active: true}
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
