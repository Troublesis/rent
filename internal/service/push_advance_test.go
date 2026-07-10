package service

import (
	"testing"
	"time"

	"github.com/troublesis/rent/config"
	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
)

// newPushServiceTestDB builds an in-memory DB with all push + payment tables
// migrated so a PushService can be wired up end to end.
func newPushServiceTestDB(t *testing.T) *repository.PushRepository {
	t.Helper()
	db := newTestDB(t)
	if err := db.AutoMigrate(&model.PushConfig{}, &model.PushTemplate{}, &model.PushLog{}); err != nil {
		t.Fatalf("migrate push models: %v", err)
	}
	return repository.NewPushRepository(db)
}

// TestQueryDuePaymentsAppliesAdvanceDays is the regression test for the
// off-by-one bug: queryDuePayments must consume AdvanceDays so a bill due in
// the near future is selected today, and excluded when the window is shorter.
func TestQueryDuePaymentsAppliesAdvanceDays(t *testing.T) {
	db := newTestDB(t)
	if err := db.AutoMigrate(&model.PushConfig{}, &model.PushTemplate{}, &model.PushLog{}); err != nil {
		t.Fatalf("migrate push models: %v", err)
	}
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	pushRepo := repository.NewPushRepository(db)
	paymentService := NewPaymentService(paymentRepo, tenantRepo)
	settingsSvc := NewSettingsService(config.Config{}, repository.NewSettingsRepository(db), pushRepo)
	pushService := NewPushService(config.Config{}, pushRepo, paymentRepo, paymentService, settingsSvc)

	room, err := NewRoomService(roomRepo, tenantRepo).CreateRoom(validRoomInput("P201", "提前推送房", "1000"))
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	tenant, err := NewTenantService(db, tenantRepo, roomRepo, paymentRepo).CheckInTenant(TenantInput{
		Name:          "提前推送租客",
		Phone:         "13800004001",
		RoomID:        room.ID,
		RentPriceYuan: "1000",
		DepositYuan:   "0",
		CheckinDate:   time.Now().AddDate(0, -6, 0),
	})
	if err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}

	// A bill due 2 days from now.
	dueInTwoDays := time.Now().AddDate(0, 0, 2)
	if _, err := paymentService.RecordPayment(PaymentInput{
		TenantID:   tenant.ID,
		AmountYuan: "1000",
		Type:       model.PaymentTypeRent,
		Paid:       false,
		PayDate:    dueInTwoDays,
	}); err != nil {
		t.Fatalf("RecordPayment returned error: %v", err)
	}

	// With AdvanceDays=3 the bill is within the window -> selected.
	setAdvanceDays(t, pushRepo, 3)
	got, err := pushService.queryDuePayments()
	if err != nil {
		t.Fatalf("queryDuePayments advance=3 returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("advance=3: len(payments) = %d, want 1", len(got))
	}

	// With AdvanceDays=1 the bill is outside the window -> not selected.
	setAdvanceDays(t, pushRepo, 1)
	got, err = pushService.queryDuePayments()
	if err != nil {
		t.Fatalf("queryDuePayments advance=1 returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("advance=1: len(payments) = %d, want 0", len(got))
	}
}

// TestQueryDuePaymentsDelayOffsetExcludesToday verifies a negative offset
// (推迟) narrows the window so a bill due today is NOT selected when the
// landlord chose to delay reminders.
func TestQueryDuePaymentsDelayOffsetExcludesToday(t *testing.T) {
	db := newTestDB(t)
	if err := db.AutoMigrate(&model.PushConfig{}, &model.PushTemplate{}, &model.PushLog{}); err != nil {
		t.Fatalf("migrate push models: %v", err)
	}
	tenantRepo := repository.NewTenantRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	pushRepo := repository.NewPushRepository(db)
	paymentService := NewPaymentService(paymentRepo, tenantRepo)
	settingsSvc := NewSettingsService(config.Config{}, repository.NewSettingsRepository(db), pushRepo)
	pushService := NewPushService(config.Config{}, pushRepo, paymentRepo, paymentService, settingsSvc)

	room, err := NewRoomService(roomRepo, tenantRepo).CreateRoom(validRoomInput("P202", "推迟推送房", "1000"))
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	tenant, err := NewTenantService(db, tenantRepo, roomRepo, paymentRepo).CheckInTenant(TenantInput{
		Name:          "推迟推送租客",
		Phone:         "13800004002",
		RoomID:        room.ID,
		RentPriceYuan: "1000",
		DepositYuan:   "0",
		CheckinDate:   time.Now().AddDate(0, -6, 0),
	})
	if err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}
	// Bill due today.
	if _, err := paymentService.RecordPayment(PaymentInput{
		TenantID:   tenant.ID,
		AmountYuan: "1000",
		Type:       model.PaymentTypeRent,
		Paid:       false,
		PayDate:    time.Now(),
	}); err != nil {
		t.Fatalf("RecordPayment returned error: %v", err)
	}

	// 推迟 3 天: window ends 3 days ago, so today's bill is excluded.
	setAdvanceDays(t, pushRepo, -3)
	got, err := pushService.queryDuePayments()
	if err != nil {
		t.Fatalf("queryDuePayments delay=-3 returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("delay=-3: len(payments) = %d, want 0 (today's bill excluded)", len(got))
	}
}

func setAdvanceDays(t *testing.T, pushRepo *repository.PushRepository, days int) {
	t.Helper()
	config, err := pushRepo.GetConfigOrCreate()
	if err != nil {
		t.Fatalf("GetConfigOrCreate: %v", err)
	}
	config.AdvanceDays = days
	config.Enabled = true
	if err := pushRepo.SaveConfig(config); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
}
