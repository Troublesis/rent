package service

import (
	"testing"
	"time"

	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
	"gorm.io/gorm"
)

func TestStatsServiceMonthlyIncomeAndOverview(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	dashboardService := NewDashboardService(roomRepo, tenantRepo, paymentRepo)
	statsService := NewStatsService(roomRepo, tenantRepo, paymentRepo, dashboardService)

	occupiedRoom := createStatsRoom(t, db, "S101", model.RoomStatusOccupied)
	createStatsRoom(t, db, "S102", model.RoomStatusVacant)
	tenant := createStatsTenant(t, db, occupiedRoom.ID, "统计租客", model.TenantStatusActive, time.Date(2026, time.January, 1, 0, 0, 0, 0, time.Local), nil)
	createStatsPayment(t, db, tenant.ID, 100000, time.Date(2026, time.January, 15, 0, 0, 0, 0, time.Local), true, false)
	createStatsPayment(t, db, tenant.ID, 200000, time.Date(2026, time.March, 15, 0, 0, 0, 0, time.Local), true, false)
	createStatsPayment(t, db, tenant.ID, 300000, time.Date(2026, time.March, 16, 0, 0, 0, 0, time.Local), false, false)
	createStatsPayment(t, db, tenant.ID, 400000, time.Date(2026, time.March, 17, 0, 0, 0, 0, time.Local), true, true)

	filter, err := NewYearStatsFilter(2026, time.Date(2026, time.May, 17, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("NewYearStatsFilter returned error: %v", err)
	}
	report, err := statsService.MonthlyIncome(filter)
	if err != nil {
		t.Fatalf("MonthlyIncome returned error: %v", err)
	}
	if len(report.Months) != 12 {
		t.Fatalf("len(report.Months) = %d, want 12", len(report.Months))
	}
	if report.Summary.TotalFen != 300000 {
		t.Fatalf("TotalFen = %d, want 300000", report.Summary.TotalFen)
	}
	if report.Summary.AverageMonthlyFen != 25000 {
		t.Fatalf("AverageMonthlyFen = %d, want 25000", report.Summary.AverageMonthlyFen)
	}
	if report.Summary.PeakMonth != "2026-03" || report.Summary.PeakMonthPaidFen != 200000 {
		t.Fatalf("peak = %s/%d, want 2026-03/200000", report.Summary.PeakMonth, report.Summary.PeakMonthPaidFen)
	}

	overview, err := statsService.Overview(filter)
	if err != nil {
		t.Fatalf("Overview returned error: %v", err)
	}
	if overview.TotalPaidFen != 300000 || overview.TotalRooms != 2 || overview.OccupiedRooms != 1 || overview.ActiveTenants != 1 {
		t.Fatalf("overview = %#v, want paid 300000, rooms 2/1, tenants 1", overview)
	}
	if overview.OccupancyRate != 0.5 {
		t.Fatalf("OccupancyRate = %v, want 0.5", overview.OccupancyRate)
	}
}

func TestStatsServiceMonthlyOccupancyApproximation(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	dashboardService := NewDashboardService(roomRepo, tenantRepo, paymentRepo)
	statsService := NewStatsService(roomRepo, tenantRepo, paymentRepo, dashboardService)

	roomA := createStatsRoom(t, db, "S201", model.RoomStatusOccupied)
	roomB := createStatsRoom(t, db, "S202", model.RoomStatusVacant)
	createStatsTenant(t, db, roomA.ID, "长期租客", model.TenantStatusActive, time.Date(2026, time.January, 1, 0, 0, 0, 0, time.Local), nil)
	checkoutDate := time.Date(2026, time.February, 20, 0, 0, 0, 0, time.Local)
	createStatsTenant(t, db, roomB.ID, "短租租客", model.TenantStatusCheckout, time.Date(2026, time.February, 15, 0, 0, 0, 0, time.Local), &checkoutDate)

	filter, err := NewDateRangeStatsFilter(time.Date(2026, time.January, 1, 0, 0, 0, 0, time.Local), time.Date(2026, time.March, 31, 0, 0, 0, 0, time.Local), time.Date(2026, time.May, 17, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("NewDateRangeStatsFilter returned error: %v", err)
	}
	report, err := statsService.MonthlyOccupancy(filter)
	if err != nil {
		t.Fatalf("MonthlyOccupancy returned error: %v", err)
	}
	if len(report.Months) != 3 {
		t.Fatalf("len(report.Months) = %d, want 3", len(report.Months))
	}
	wantOccupied := []int{1, 2, 1}
	for i, want := range wantOccupied {
		if report.Months[i].OccupiedRooms != want {
			t.Fatalf("month %d occupied = %d, want %d", i, report.Months[i].OccupiedRooms, want)
		}
	}
	if !report.Approximate || report.Note == "" {
		t.Fatalf("report should be marked approximate with note: %#v", report)
	}
}

func createStatsRoom(t *testing.T, db interface {
	Create(value interface{}) *gorm.DB
}, roomNo string, status string) model.Room {
	t.Helper()
	room := model.Room{RoomNo: roomNo, Title: roomNo + " 房源", RentType: model.RentTypeMonthly, RentPrice: 100000, PaymentTerms: model.PaymentTerms1M1D, Deposit: 100000, Status: status}
	if err := db.Create(&room).Error; err != nil {
		t.Fatalf("create room: %v", err)
	}
	return room
}

func createStatsTenant(t *testing.T, db interface {
	Create(value interface{}) *gorm.DB
}, roomID uint, name string, status string, checkinDate time.Time, checkoutDate *time.Time) model.Tenant {
	t.Helper()
	tenant := model.Tenant{Name: name, Phone: "13800002000", RoomID: roomID, CheckinDate: checkinDate, CheckoutDate: checkoutDate, RentPrice: 100000, RentType: model.RentTypeMonthly, PaymentTerms: model.PaymentTerms1M1D, Status: status}
	if err := db.Create(&tenant).Error; err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	return tenant
}

func createStatsPayment(t *testing.T, db interface {
	Create(value interface{}) *gorm.DB
}, tenantID uint, amount int, payDate time.Time, paid bool, excluded bool) model.Payment {
	t.Helper()
	payment := model.Payment{TenantID: tenantID, Amount: amount, Type: model.PaymentTypeRent, Paid: paid, PayDate: payDate, Excluded: excluded}
	if err := db.Create(&payment).Error; err != nil {
		t.Fatalf("create payment: %v", err)
	}
	return payment
}
