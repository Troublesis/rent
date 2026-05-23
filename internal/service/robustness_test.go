package service

import (
	"testing"
	"time"

	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
)

// TestDaysBetweenDSTBoundary pins down the calendar-day calculation across a
// DST transition. In US/Eastern, 2026-03-08 is a 23h day (clocks jump
// forward) and 2026-11-01 is a 25h day. The old `Hours()/24` implementation
// would have returned 6 instead of 7 for the spring-forward case.
func TestDaysBetweenDSTBoundary(t *testing.T) {
	eastern, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skipf("zoneinfo unavailable: %v", err)
	}
	cases := []struct {
		name  string
		start time.Time
		end   time.Time
		want  int
	}{
		{
			name:  "spring forward (23h day)",
			start: time.Date(2026, 3, 5, 0, 0, 0, 0, eastern),
			end:   time.Date(2026, 3, 12, 0, 0, 0, 0, eastern),
			want:  7,
		},
		{
			name:  "fall back (25h day)",
			start: time.Date(2026, 10, 30, 0, 0, 0, 0, eastern),
			end:   time.Date(2026, 11, 6, 0, 0, 0, 0, eastern),
			want:  7,
		},
		{
			name:  "same day → 0",
			start: time.Date(2026, 5, 1, 9, 0, 0, 0, eastern),
			end:   time.Date(2026, 5, 1, 23, 0, 0, 0, eastern),
			want:  0,
		},
		{
			name:  "single calendar day",
			start: time.Date(2026, 5, 1, 0, 0, 0, 0, eastern),
			end:   time.Date(2026, 5, 2, 0, 0, 0, 0, eastern),
			want:  1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := daysBetween(tc.start, tc.end)
			if got != tc.want {
				t.Errorf("daysBetween(%s, %s) = %d, want %d", tc.start.Format(time.RFC3339), tc.end.Format(time.RFC3339), got, tc.want)
			}
		})
	}
}

// TestMidMonthDailyRentReceivableClampedToCheckin documents the existing
// behaviour where a daily-rent tenant who checks in mid-window contributes
// rent only from the check-in date onwards, even when the projection window
// starts earlier.
func TestMidMonthDailyRentReceivableClampedToCheckin(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	dashboardService := NewDashboardService(roomRepo, tenantRepo, paymentRepo)

	room := &model.Room{RoomNo: "X1", Title: "短租房", RentType: model.RentTypeDaily, RentPrice: 10000, Status: model.RoomStatusOccupied}
	if err := db.Create(room).Error; err != nil {
		t.Fatalf("create room: %v", err)
	}
	checkin := time.Date(2026, 5, 15, 0, 0, 0, 0, time.Local)
	tenant := &model.Tenant{
		Name: "测试", Phone: "13800000001", RoomID: room.ID,
		CheckinDate: checkin, RentPrice: 10000,
		RentType: model.RentTypeDaily, PaymentTerms: model.PaymentTerms1M1D, Status: model.TenantStatusActive,
	}
	if err := db.Create(tenant).Error; err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.Local)
	summary, err := dashboardService.Summary(now)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	// May 15 → May 31 = 17 days (clamped to checkin, runs through month end).
	want := 10000 * 17
	if summary.CurrentMonthReceivable != want {
		t.Errorf("CurrentMonthReceivable = %d, want %d", summary.CurrentMonthReceivable, want)
	}
}

// TestMonthlyIncomeRangeAcrossYearBoundary verifies that the SQLite-backed
// monthly aggregator handles ranges that cross December → January correctly.
// We pin the timezone to Asia/Shanghai (the production default in
// `.env.example` and `cmd/server/main.go`) so the test isn't sensitive to the
// developer's machine clock.
func TestMonthlyIncomeRangeAcrossYearBoundary(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Skipf("zoneinfo unavailable: %v", err)
	}
	original := time.Local
	time.Local = loc
	t.Cleanup(func() { time.Local = original })

	db := newTestDB(t)
	paymentRepo := repository.NewPaymentRepository(db)

	room := &model.Room{RoomNo: "Y1", Title: "整年", RentType: model.RentTypeMonthly, RentPrice: 100000, Status: model.RoomStatusOccupied}
	if err := db.Create(room).Error; err != nil {
		t.Fatalf("create room: %v", err)
	}
	tenant := &model.Tenant{
		Name: "跨年", Phone: "13900000001", RoomID: room.ID,
		CheckinDate: time.Date(2025, 1, 1, 0, 0, 0, 0, loc), RentPrice: 100000,
		RentType: model.RentTypeMonthly, PaymentTerms: model.PaymentTerms1M1D, Status: model.TenantStatusActive,
	}
	if err := db.Create(tenant).Error; err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	payments := []model.Payment{
		{TenantID: tenant.ID, Amount: 100000, Type: model.PaymentTypeRent, Paid: true, PayDate: time.Date(2025, 12, 1, 12, 0, 0, 0, loc)},
		{TenantID: tenant.ID, Amount: 100000, Type: model.PaymentTypeRent, Paid: true, PayDate: time.Date(2026, 1, 5, 12, 0, 0, 0, loc)},
		{TenantID: tenant.ID, Amount: 100000, Type: model.PaymentTypeRent, Paid: true, PayDate: time.Date(2026, 2, 5, 12, 0, 0, 0, loc)},
	}
	if err := db.Create(&payments).Error; err != nil {
		t.Fatalf("create payments: %v", err)
	}

	start := time.Date(2025, 12, 1, 0, 0, 0, 0, loc)
	end := time.Date(2026, 3, 1, 0, 0, 0, 0, loc)
	rows, err := paymentRepo.MonthlyIncomeRange(start, end)
	if err != nil {
		t.Fatalf("monthly income range: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 month rows, got %d (%+v)", len(rows), rows)
	}
	wantYears := []int{2025, 2026, 2026}
	wantMonths := []int{12, 1, 2}
	for i, row := range rows {
		if row.Year != wantYears[i] || row.Month != wantMonths[i] {
			t.Errorf("row %d = %v, want year=%d month=%d", i, row, wantYears[i], wantMonths[i])
		}
		if row.Total != 100000 {
			t.Errorf("row %d total = %d, want 100000", i, row.Total)
		}
	}
}

// TestTogglePaidClearsExclusion regression test: marking a payment as paid
// must clear the excluded flag and exclusion note so that the same record
// cannot simultaneously be "已付款" and "已排除".
func TestTogglePaidClearsExclusion(t *testing.T) {
	db := newTestDB(t)
	paymentRepo := repository.NewPaymentRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentService := NewPaymentService(paymentRepo, tenantRepo)

	room := &model.Room{RoomNo: "Z1", Title: "排除测试", RentType: model.RentTypeMonthly, RentPrice: 50000, Status: model.RoomStatusOccupied}
	if err := db.Create(room).Error; err != nil {
		t.Fatalf("create room: %v", err)
	}
	tenant := &model.Tenant{
		Name: "排除", Phone: "13900000002", RoomID: room.ID,
		CheckinDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local), RentPrice: 50000,
		RentType: model.RentTypeMonthly, PaymentTerms: model.PaymentTerms1M1D, Status: model.TenantStatusCheckout,
	}
	if err := db.Create(tenant).Error; err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	payment := &model.Payment{
		TenantID: tenant.ID, Amount: 50000, Type: model.PaymentTypeRent,
		Paid: false, PayDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.Local),
	}
	if err := db.Create(payment).Error; err != nil {
		t.Fatalf("create payment: %v", err)
	}

	if err := paymentService.SetExcluded(payment.ID, true, "押金抵扣"); err != nil {
		t.Fatalf("set excluded: %v", err)
	}
	if err := paymentService.TogglePaid(payment.ID); err != nil {
		t.Fatalf("toggle paid: %v", err)
	}
	reloaded, err := paymentRepo.GetPayment(payment.ID)
	if err != nil {
		t.Fatalf("get payment: %v", err)
	}
	if !reloaded.Paid {
		t.Errorf("expected payment marked paid")
	}
	if reloaded.Excluded {
		t.Errorf("expected excluded flag cleared after marking paid")
	}
	if reloaded.ExclusionNote != "" {
		t.Errorf("expected exclusion note cleared, got %q", reloaded.ExclusionNote)
	}
}
