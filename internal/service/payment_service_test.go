package service

import (
	"testing"
	"time"

	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
)

func TestPaymentToggleAndMonthlyIncome(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)
	paymentService := NewPaymentService(paymentRepo, tenantRepo)

	room, err := roomService.CreateRoom(RoomInput{RoomNo: "B201", Title: "朝南单间", PriceYuan: "1200", DepositYuan: "1200", Status: model.RoomStatusVacant})
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	tenant, err := tenantService.CheckInTenant(TenantInput{Name: "钱七", Phone: "13800000004", RoomID: room.ID, RentPriceYuan: "1200", DepositYuan: "1200"})
	if err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}

	paidPayment, err := paymentService.RecordPayment(PaymentInput{
		TenantID:   tenant.ID,
		AmountYuan: "1200",
		Type:       model.PaymentTypeRent,
		Paid:       true,
		PayDate:    time.Date(2026, time.May, 3, 0, 0, 0, 0, time.Local),
	})
	if err != nil {
		t.Fatalf("RecordPayment paid returned error: %v", err)
	}
	if _, err := paymentService.RecordPayment(PaymentInput{
		TenantID:   tenant.ID,
		AmountYuan: "80",
		Type:       model.PaymentTypeWater,
		Paid:       false,
		PayDate:    time.Date(2026, time.May, 4, 0, 0, 0, 0, time.Local),
	}); err != nil {
		t.Fatalf("RecordPayment unpaid returned error: %v", err)
	}

	rows, err := paymentService.MonthlyIncome(2026)
	if err != nil {
		t.Fatalf("MonthlyIncome returned error: %v", err)
	}
	if len(rows) != 1 || rows[0].Month != 5 || rows[0].Total != 120000 {
		t.Fatalf("rows = %#v, want May total 120000", rows)
	}

	if err := paymentService.TogglePaid(paidPayment.ID); err != nil {
		t.Fatalf("TogglePaid returned error: %v", err)
	}
	updatedPayment, err := paymentRepo.GetPayment(paidPayment.ID)
	if err != nil {
		t.Fatalf("GetPayment returned error: %v", err)
	}
	if updatedPayment.Paid {
		t.Fatal("payment should be toggled to unpaid")
	}
}
