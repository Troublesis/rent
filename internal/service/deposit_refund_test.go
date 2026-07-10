package service

import (
	"testing"
	"time"

	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
)

// TestCheckOutCreatesDepositRefund verifies the core feature: checking out a
// tenant who paid a deposit spawns an unpaid deposit_refund record (negative
// amount) so the landlord is reminded to return the deposit.
func TestCheckOutCreatesDepositRefund(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	tenantService := NewTenantService(db, tenantRepo, roomRepo, paymentRepo)

	room, err := NewRoomService(roomRepo, tenantRepo).CreateRoom(validRoomInput("R201", "押金房", "1500"))
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	tenant, err := tenantService.CheckInTenant(TenantInput{
		Name:          "押金租客",
		Phone:         "13800003001",
		RoomID:        room.ID,
		RentPriceYuan: "1500",
		DepositYuan:   "1500",
	})
	if err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}
	if err := tenantService.CheckOutTenant(tenant.ID); err != nil {
		t.Fatalf("CheckOutTenant returned error: %v", err)
	}

	refunds, err := paymentRepo.ListPayments(repository.PaymentFilter{
		TenantID: tenant.ID,
		Type:     model.PaymentTypeDepositRefund,
	})
	if err != nil {
		t.Fatalf("ListPayments returned error: %v", err)
	}
	if len(refunds) != 1 {
		t.Fatalf("len(refunds) = %d, want 1", len(refunds))
	}
	r := refunds[0]
	if r.Amount != -150000 {
		t.Fatalf("refund amount = %d, want -150000", r.Amount)
	}
	if r.Paid {
		t.Fatal("refund should be unpaid immediately after checkout")
	}
	if r.Type != model.PaymentTypeDepositRefund {
		t.Fatalf("refund type = %q, want %q", r.Type, model.PaymentTypeDepositRefund)
	}
}

// TestCheckOutDepositRefundIdempotent ensures a second checkout attempt does
// not create a duplicate refund row.
func TestCheckOutDepositRefundIdempotent(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	tenantService := NewTenantService(db, tenantRepo, roomRepo, paymentRepo)

	room, err := NewRoomService(roomRepo, tenantRepo).CreateRoom(validRoomInput("R202", "重复退租房", "1000"))
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	tenant, err := tenantService.CheckInTenant(TenantInput{
		Name:          "重复租客",
		Phone:         "13800003002",
		RoomID:        room.ID,
		RentPriceYuan: "1000",
		DepositYuan:   "1000",
	})
	if err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}
	if err := tenantService.CheckOutTenant(tenant.ID); err != nil {
		t.Fatalf("first CheckOutTenant returned error: %v", err)
	}
	// Second checkout fails (tenant no longer active), so no duplicate path —
	// but verify the count is still one regardless.
	_ = tenantService.CheckOutTenant(tenant.ID)

	count, err := paymentRepo.HasDepositRefund(tenant.ID)
	if err != nil {
		t.Fatalf("HasDepositRefund returned error: %v", err)
	}
	if !count {
		t.Fatal("HasDepositRefund = false, want true")
	}
	refunds, err := paymentRepo.ListPayments(repository.PaymentFilter{
		TenantID: tenant.ID,
		Type:     model.PaymentTypeDepositRefund,
	})
	if err != nil {
		t.Fatalf("ListPayments returned error: %v", err)
	}
	if len(refunds) != 1 {
		t.Fatalf("len(refunds) = %d, want 1 (no duplicate)", len(refunds))
	}
}

// TestCheckOutNoDepositCreatesNoRefund verifies a tenant with zero deposit
// produces no refund record on checkout.
func TestCheckOutNoDepositCreatesNoRefund(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	tenantService := NewTenantService(db, tenantRepo, roomRepo, paymentRepo)

	room, err := NewRoomService(roomRepo, tenantRepo).CreateRoom(validRoomInput("R203", "无押金房", "1000"))
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	tenant, err := tenantService.CheckInTenant(TenantInput{
		Name:          "无押金租客",
		Phone:         "13800003003",
		RoomID:        room.ID,
		RentPriceYuan: "1000",
		DepositYuan:   "0",
	})
	if err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}
	if err := tenantService.CheckOutTenant(tenant.ID); err != nil {
		t.Fatalf("CheckOutTenant returned error: %v", err)
	}
	refunds, err := paymentRepo.ListPayments(repository.PaymentFilter{
		TenantID: tenant.ID,
		Type:     model.PaymentTypeDepositRefund,
	})
	if err != nil {
		t.Fatalf("ListPayments returned error: %v", err)
	}
	if len(refunds) != 0 {
		t.Fatalf("len(refunds) = %d, want 0 for zero-deposit tenant", len(refunds))
	}
}

// TestDepositRefundDeductsFromIncome verifies the money semantics: a refunded
// (paid) deposit deducts from total collected income via the shared SUM query.
func TestDepositRefundDeductsFromIncome(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	tenantService := NewTenantService(db, tenantRepo, roomRepo, paymentRepo)

	room, err := NewRoomService(roomRepo, tenantRepo).CreateRoom(validRoomInput("R204", "押金扣除房", "1000"))
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	tenant, err := tenantService.CheckInTenant(TenantInput{
		Name:          "扣除租客",
		Phone:         "13800003004",
		RoomID:        room.ID,
		RentPriceYuan: "1000",
		DepositYuan:   "1000",
	})
	if err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}
	// Collected rent of 1000.00 => income 100000 fen.
	if _, err := NewPaymentService(paymentRepo, tenantRepo).RecordPayment(PaymentInput{
		TenantID:   tenant.ID,
		AmountYuan: "1000",
		Type:       model.PaymentTypeRent,
		Paid:       true,
		PayDate:    time.Date(2026, time.June, 5, 0, 0, 0, 0, time.Local),
	}); err != nil {
		t.Fatalf("RecordPayment returned error: %v", err)
	}
	if err := tenantService.CheckOutTenant(tenant.ID); err != nil {
		t.Fatalf("CheckOutTenant returned error: %v", err)
	}

	now := time.Now()
	// Before refunding: income = 100000 (rent only); refund is unpaid so not in SUM.
	summary, err := paymentRepo.SummarizePayments(repository.PaymentFilter{}, now)
	if err != nil {
		t.Fatalf("SummarizePayments before refund returned error: %v", err)
	}
	if summary.TotalPaidAmount != 100000 {
		t.Fatalf("TotalPaidAmount before refund = %d, want 100000", summary.TotalPaidAmount)
	}
	// Unpaid refund must NOT inflate the receivable (待收) total.
	if summary.TotalUnpaidAmount != 0 {
		t.Fatalf("TotalUnpaidAmount before refund = %d, want 0 (refund excluded)", summary.TotalUnpaidAmount)
	}

	// Locate the refund record and mark it paid (= deposit returned to tenant).
	refunds, err := paymentRepo.ListPayments(repository.PaymentFilter{
		TenantID: tenant.ID,
		Type:     model.PaymentTypeDepositRefund,
	})
	if err != nil || len(refunds) != 1 {
		t.Fatalf("expected 1 refund, got %v / %d", err, len(refunds))
	}
	paymentService := NewPaymentService(paymentRepo, tenantRepo)
	if err := paymentService.TogglePaid(refunds[0].ID); err != nil {
		t.Fatalf("TogglePaid refund returned error: %v", err)
	}

	// After refunding: income = 100000 - 100000 = 0 (deposit returned).
	summary, err = paymentRepo.SummarizePayments(repository.PaymentFilter{}, now)
	if err != nil {
		t.Fatalf("SummarizePayments after refund returned error: %v", err)
	}
	if summary.TotalPaidAmount != 0 {
		t.Fatalf("TotalPaidAmount after refund = %d, want 0 (deposit deducted)", summary.TotalPaidAmount)
	}
}
