package service

import (
	"testing"
	"time"

	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
)

func TestDashboardSummaryReceivables(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)
	paymentService := NewPaymentService(paymentRepo, tenantRepo)
	dashboardService := NewDashboardService(roomRepo, tenantRepo, paymentRepo)

	room, err := roomService.CreateRoom(validRoomInput("D101", "月租房", "1000"))
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	tenant, err := tenantService.CheckInTenant(TenantInput{Name: "吴十", Phone: "13800000010", RoomID: room.ID, CheckinDate: time.Date(2026, time.May, 1, 0, 0, 0, 0, time.Local)})
	if err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}
	if _, err := paymentService.RecordPayment(PaymentInput{TenantID: tenant.ID, AmountYuan: "1000", Type: model.PaymentTypeRent, Paid: true, PayDate: time.Date(2026, time.May, 1, 0, 0, 0, 0, time.Local)}); err != nil {
		t.Fatalf("RecordPayment paid returned error: %v", err)
	}
	excluded, err := paymentService.RecordPayment(PaymentInput{TenantID: tenant.ID, AmountYuan: "1000", Type: model.PaymentTypeRent, Paid: false, PayDate: time.Date(2026, time.June, 1, 0, 0, 0, 0, time.Local)})
	if err != nil {
		t.Fatalf("RecordPayment unpaid returned error: %v", err)
	}
	if err := paymentService.SetExcluded(excluded.ID, true, "协商减免"); err != nil {
		t.Fatalf("SetExcluded returned error: %v", err)
	}

	summary, err := dashboardService.Summary(time.Date(2026, time.July, 1, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("Summary returned error: %v", err)
	}
	if summary.UnpaidAmount != 100000 {
		t.Fatalf("UnpaidAmount = %d, want 100000", summary.UnpaidAmount)
	}
	if summary.CurrentMonthReceivable != 100000 {
		t.Fatalf("CurrentMonthReceivable = %d, want 100000", summary.CurrentMonthReceivable)
	}
	if summary.NextSixMonthsReceivable != 600000 {
		t.Fatalf("NextSixMonthsReceivable = %d, want 600000", summary.NextSixMonthsReceivable)
	}
}

func TestDashboardSummaryCountsOnlyDatedExpiredActiveTenants(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)
	dashboardService := NewDashboardService(roomRepo, tenantRepo, paymentRepo)

	expiredRoom, err := roomService.CreateRoom(validRoomInput("D103", "到期房", "1000"))
	if err != nil {
		t.Fatalf("CreateRoom expired returned error: %v", err)
	}
	if _, err := tenantService.CheckInTenant(TenantInput{Name: "到期租客", Phone: "13800000017", RoomID: expiredRoom.ID, CheckinDate: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.Local), LeaseEndDate: time.Date(2026, time.May, 1, 0, 0, 0, 0, time.Local)}); err != nil {
		t.Fatalf("CheckInTenant expired returned error: %v", err)
	}
	longTermRoom, err := roomService.CreateRoom(validRoomInput("D104", "长租房", "1000"))
	if err != nil {
		t.Fatalf("CreateRoom long term returned error: %v", err)
	}
	if _, err := tenantService.CheckInTenant(TenantInput{Name: "长租租客", Phone: "13800000018", RoomID: longTermRoom.ID, CheckinDate: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.Local)}); err != nil {
		t.Fatalf("CheckInTenant long term returned error: %v", err)
	}

	summary, err := dashboardService.Summary(time.Date(2026, time.May, 17, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("Summary returned error: %v", err)
	}
	if summary.OverdueCheckoutTenants != 1 {
		t.Fatalf("OverdueCheckoutTenants = %d, want 1", summary.OverdueCheckoutTenants)
	}
}

func TestDashboardProjectionDetailsShareSummaryTotals(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)
	dashboardService := NewDashboardService(roomRepo, tenantRepo, paymentRepo)

	room, err := roomService.CreateRoom(validRoomInput("D105", "预测房", "1000"))
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	if _, err := tenantService.CheckInTenant(TenantInput{Name: "预测租客", Phone: "13800000019", RoomID: room.ID, CheckinDate: time.Date(2026, time.May, 1, 0, 0, 0, 0, time.Local)}); err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}

	now := time.Date(2026, time.May, 17, 12, 0, 0, 0, time.Local)
	summary, err := dashboardService.Summary(now)
	if err != nil {
		t.Fatalf("Summary returned error: %v", err)
	}
	projection, err := dashboardService.Projection("6months", now)
	if err != nil {
		t.Fatalf("Projection returned error: %v", err)
	}
	if projection.Total != summary.NextSixMonthsReceivable {
		t.Fatalf("projection total = %d, want summary %d", projection.Total, summary.NextSixMonthsReceivable)
	}
	if len(projection.Months) != 6 {
		t.Fatalf("len(projection.Months) = %d, want 6", len(projection.Months))
	}
}

func TestDashboardDailyProjectionUsesCalendarDays(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)
	dashboardService := NewDashboardService(roomRepo, tenantRepo, paymentRepo)

	input := validRoomInput("D102", "日租房", "80")
	input.RentType = model.RentTypeDaily
	room, err := roomService.CreateRoom(input)
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	if _, err := tenantService.CheckInTenant(TenantInput{Name: "郑十", Phone: "13800000011", RoomID: room.ID, CheckinDate: time.Date(2026, time.May, 10, 0, 0, 0, 0, time.Local)}); err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}
	summary, err := dashboardService.Summary(time.Date(2026, time.May, 17, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("Summary returned error: %v", err)
	}
	if summary.UnpaidAmount != 64000 {
		t.Fatalf("UnpaidAmount = %d, want 64000", summary.UnpaidAmount)
	}
	if summary.CurrentMonthReceivable != 176000 {
		t.Fatalf("CurrentMonthReceivable = %d, want 176000", summary.CurrentMonthReceivable)
	}
}
