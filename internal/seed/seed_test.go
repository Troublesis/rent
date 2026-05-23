package seed_test

import (
	"strings"
	"testing"
	"time"

	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
	"github.com/troublesis/rent/internal/seed"
	"github.com/troublesis/rent/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSeedDB(t *testing.T) *gorm.DB {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.AutoMigrate(&model.Room{}, &model.RoomMedia{}, &model.Tenant{}, &model.Payment{}, &model.AppSetting{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

// TestSeedRunConsistency exercises the seeder on a clean in-memory database
// and verifies the resulting fixture is self-consistent — every active
// tenant's room is occupied, every checkout tenant's room is vacant, and
// payments excluded from the books belong only to checkout tenants. It also
// runs the dashboard service against the seeded data to ensure all summary
// numbers are finite and non-negative.
func TestSeedRunConsistency(t *testing.T) {
	db := newSeedDB(t)
	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.Local)
	stats, err := seed.Run(db, now)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	if stats.Rooms == 0 || stats.Tenants == 0 || stats.Payments == 0 {
		t.Fatalf("seeded counts look empty: %+v", stats)
	}

	var rooms []model.Room
	if err := db.Find(&rooms).Error; err != nil {
		t.Fatalf("list rooms: %v", err)
	}
	roomByID := make(map[uint]model.Room, len(rooms))
	for _, r := range rooms {
		roomByID[r.ID] = r
	}

	var tenants []model.Tenant
	if err := db.Find(&tenants).Error; err != nil {
		t.Fatalf("list tenants: %v", err)
	}
	for _, tenant := range tenants {
		room, ok := roomByID[tenant.RoomID]
		if !ok {
			t.Errorf("tenant %s references missing room %d", tenant.Name, tenant.RoomID)
			continue
		}
		switch tenant.Status {
		case model.TenantStatusActive:
			if room.Status != model.RoomStatusOccupied {
				t.Errorf("active tenant %s has room %s in status %q, want occupied", tenant.Name, room.RoomNo, room.Status)
			}
		case model.TenantStatusCheckout:
			if room.Status == model.RoomStatusOccupied {
				t.Errorf("checkout tenant %s has room %s still occupied", tenant.Name, room.RoomNo)
			}
			if tenant.CheckoutDate == nil {
				t.Errorf("checkout tenant %s missing checkout date", tenant.Name)
			}
		}
	}

	var excluded []model.Payment
	if err := db.Where("excluded = ?", true).Preload("Tenant").Find(&excluded).Error; err != nil {
		t.Fatalf("list excluded payments: %v", err)
	}
	for _, payment := range excluded {
		if payment.Tenant.Status != model.TenantStatusCheckout {
			t.Errorf("excluded payment %d belongs to non-checkout tenant %s (status %q)", payment.ID, payment.Tenant.Name, payment.Tenant.Status)
		}
	}

	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	dashboard := service.NewDashboardService(roomRepo, tenantRepo, paymentRepo)
	summary, err := dashboard.Summary(now)
	if err != nil {
		t.Fatalf("dashboard summary: %v", err)
	}
	values := map[string]int{
		"CurrentMonthIncome":         summary.CurrentMonthIncome,
		"UnpaidAmount":               summary.UnpaidAmount,
		"CurrentMonthReceivable":     summary.CurrentMonthReceivable,
		"NextSixMonthsReceivable":    summary.NextSixMonthsReceivable,
		"NextTwelveMonthsReceivable": summary.NextTwelveMonthsReceivable,
	}
	for name, value := range values {
		if value < 0 {
			t.Errorf("summary.%s = %d, want non-negative", name, value)
		}
	}
	if summary.TotalRooms <= 0 || summary.ActiveTenants <= 0 {
		t.Errorf("expected non-zero summary counts, got rooms=%d active=%d", summary.TotalRooms, summary.ActiveTenants)
	}
}
