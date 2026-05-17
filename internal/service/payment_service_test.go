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

	room, err := roomService.CreateRoom(validRoomInput("B201", "朝南单间", "1200"))
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

func TestPaymentExclusionRemovesUnpaidAggregate(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)
	paymentService := NewPaymentService(paymentRepo, tenantRepo)

	room, err := roomService.CreateRoom(validRoomInput("B202", "西向单间", "1000"))
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	tenant, err := tenantService.CheckInTenant(TenantInput{Name: "孙八", Phone: "13800000005", RoomID: room.ID})
	if err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}
	payment, err := paymentService.RecordPayment(PaymentInput{TenantID: tenant.ID, AmountYuan: "1000", Type: model.PaymentTypeRent, Paid: false})
	if err != nil {
		t.Fatalf("RecordPayment returned error: %v", err)
	}
	before, err := paymentService.SumUnpaid()
	if err != nil {
		t.Fatalf("SumUnpaid returned error: %v", err)
	}
	if before != 100000 {
		t.Fatalf("before = %d, want 100000", before)
	}
	if err := paymentService.SetExcluded(payment.ID, true, "租客已搬离"); err != nil {
		t.Fatalf("SetExcluded returned error: %v", err)
	}
	after, err := paymentService.SumUnpaid()
	if err != nil {
		t.Fatalf("SumUnpaid after returned error: %v", err)
	}
	if after != 0 {
		t.Fatalf("after = %d, want 0", after)
	}
	updatedPayment, err := paymentRepo.GetPayment(payment.ID)
	if err != nil {
		t.Fatalf("GetPayment returned error: %v", err)
	}
	if !updatedPayment.Excluded || updatedPayment.ExclusionNote != "租客已搬离" {
		t.Fatalf("updated payment exclusion = %v/%q", updatedPayment.Excluded, updatedPayment.ExclusionNote)
	}
	if err := paymentService.TogglePaid(payment.ID); err != nil {
		t.Fatalf("TogglePaid returned error: %v", err)
	}
	paidPayment, err := paymentRepo.GetPayment(payment.ID)
	if err != nil {
		t.Fatalf("GetPayment paid returned error: %v", err)
	}
	if paidPayment.Excluded || paidPayment.ExclusionNote != "" {
		t.Fatalf("paid payment exclusion = %v/%q, want cleared", paidPayment.Excluded, paidPayment.ExclusionNote)
	}
}

func TestRentDueScheduleMonthlyQuarterlyAndDaily(t *testing.T) {
	monthlyTenant := model.Tenant{
		ID:           1,
		Status:       model.TenantStatusActive,
		CheckinDate:  time.Date(2026, time.January, 5, 0, 0, 0, 0, time.Local),
		RentPrice:    100000,
		RentType:     model.RentTypeMonthly,
		PaymentTerms: model.PaymentTerms3M1D,
	}
	monthlyDues := RentDueSchedule(monthlyTenant, time.Date(2026, time.July, 5, 12, 0, 0, 0, time.Local))
	if len(monthlyDues) != 3 {
		t.Fatalf("len(monthlyDues) = %d, want 3", len(monthlyDues))
	}
	if monthlyDues[0].Amount != 300000 || monthlyDues[1].Amount != 300000 || monthlyDues[2].Amount != 300000 {
		t.Fatalf("monthly due amounts = %#v, want 300000 each", monthlyDues)
	}
	if !monthlyDues[2].DueDate.Equal(time.Date(2026, time.July, 5, 0, 0, 0, 0, time.Local)) {
		t.Fatalf("last due date = %v, want 2026-07-05", monthlyDues[2].DueDate)
	}

	dailyTenant := model.Tenant{
		ID:          2,
		Status:      model.TenantStatusActive,
		CheckinDate: time.Date(2026, time.May, 15, 0, 0, 0, 0, time.Local),
		RentPrice:   8000,
		RentType:    model.RentTypeDaily,
	}
	dailyDues := RentDueSchedule(dailyTenant, time.Date(2026, time.May, 17, 12, 0, 0, 0, time.Local))
	if len(dailyDues) != 3 {
		t.Fatalf("len(dailyDues) = %d, want 3", len(dailyDues))
	}
	if dailyDues[0].Amount != 8000 || dailyDues[2].Amount != 8000 {
		t.Fatalf("daily due amounts = %#v, want 8000 each", dailyDues)
	}
}

func TestGenerateDueRecordsCreatesMissingMonthlyRecords(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)
	paymentService := NewPaymentService(paymentRepo, tenantRepo)

	room, err := roomService.CreateRoom(validRoomInput("B203", "月租单间", "1000"))
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	tenant, err := tenantService.CheckInTenant(TenantInput{Name: "何十", Phone: "13800000014", RoomID: room.ID, CheckinDate: time.Date(2026, time.January, 5, 0, 0, 0, 0, time.Local)})
	if err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}
	created, err := paymentService.GenerateDueRecords(tenant.ID, time.Date(2026, time.March, 5, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("GenerateDueRecords returned error: %v", err)
	}
	if created != 3 {
		t.Fatalf("created = %d, want 3", created)
	}
	payments, err := paymentRepo.ListPayments(repository.PaymentFilter{TenantID: tenant.ID, Type: model.PaymentTypeRent})
	if err != nil {
		t.Fatalf("ListPayments returned error: %v", err)
	}
	if len(payments) != 3 {
		t.Fatalf("len(payments) = %d, want 3", len(payments))
	}
	for _, payment := range payments {
		if !payment.AutoGenerated || payment.Paid || payment.Excluded || payment.Amount != 100000 || payment.Type != model.PaymentTypeRent {
			t.Fatalf("generated payment = %#v, want unpaid auto rent amount 100000", payment)
		}
	}
}

func TestGenerateDueRecordsSkipsExistingRecordsAndIsIdempotent(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)
	paymentService := NewPaymentService(paymentRepo, tenantRepo)

	room, err := roomService.CreateRoom(validRoomInput("B204", "免租单间", "1000"))
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	tenant, err := tenantService.CheckInTenant(TenantInput{Name: "吕十", Phone: "13800000015", RoomID: room.ID, CheckinDate: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.Local)})
	if err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}
	manualPayment, err := paymentService.RecordPayment(PaymentInput{TenantID: tenant.ID, AmountYuan: "1000", Type: model.PaymentTypeRent, Paid: false, PayDate: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.Local)})
	if err != nil {
		t.Fatalf("RecordPayment returned error: %v", err)
	}
	if err := paymentService.SetExcluded(manualPayment.ID, true, "已线下处理"); err != nil {
		t.Fatalf("SetExcluded returned error: %v", err)
	}
	created, err := paymentService.GenerateDueRecords(tenant.ID, time.Date(2026, time.February, 1, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("GenerateDueRecords returned error: %v", err)
	}
	if created != 1 {
		t.Fatalf("created = %d, want 1", created)
	}
	createdAgain, err := paymentService.GenerateDueRecords(tenant.ID, time.Date(2026, time.February, 1, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("GenerateDueRecords second call returned error: %v", err)
	}
	if createdAgain != 0 {
		t.Fatalf("createdAgain = %d, want 0", createdAgain)
	}
	payments, err := paymentRepo.ListPayments(repository.PaymentFilter{TenantID: tenant.ID, Type: model.PaymentTypeRent})
	if err != nil {
		t.Fatalf("ListPayments returned error: %v", err)
	}
	if len(payments) != 2 {
		t.Fatalf("len(payments) = %d, want 2", len(payments))
	}
}

func TestGenerateDueRecordsCreatesDailyRecords(t *testing.T) {
	db := newTestDB(t)
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	roomService := NewRoomService(roomRepo, tenantRepo)
	tenantService := NewTenantService(db, tenantRepo, roomRepo)
	paymentService := NewPaymentService(paymentRepo, tenantRepo)

	input := validRoomInput("B205", "日租单间", "80")
	input.RentType = model.RentTypeDaily
	room, err := roomService.CreateRoom(input)
	if err != nil {
		t.Fatalf("CreateRoom returned error: %v", err)
	}
	tenant, err := tenantService.CheckInTenant(TenantInput{Name: "施十", Phone: "13800000016", RoomID: room.ID, CheckinDate: time.Date(2026, time.May, 15, 0, 0, 0, 0, time.Local)})
	if err != nil {
		t.Fatalf("CheckInTenant returned error: %v", err)
	}
	created, err := paymentService.GenerateDueRecords(tenant.ID, time.Date(2026, time.May, 17, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("GenerateDueRecords returned error: %v", err)
	}
	if created != 3 {
		t.Fatalf("created = %d, want 3", created)
	}
	payments, err := paymentRepo.ListPayments(repository.PaymentFilter{TenantID: tenant.ID, Type: model.PaymentTypeRent})
	if err != nil {
		t.Fatalf("ListPayments returned error: %v", err)
	}
	for _, payment := range payments {
		if payment.Amount != 8000 {
			t.Fatalf("payment amount = %d, want 8000", payment.Amount)
		}
	}
}

func TestNextDueForTenant(t *testing.T) {
	tenant := model.Tenant{
		CheckinDate:  time.Date(2026, time.May, 5, 0, 0, 0, 0, time.Local),
		RentType:     model.RentTypeMonthly,
		PaymentTerms: model.PaymentTerms3M1D,
		Payments: []model.Payment{{
			Type:    model.PaymentTypeRent,
			Paid:    true,
			PayDate: time.Date(2026, time.May, 5, 0, 0, 0, 0, time.Local),
		}},
	}
	dueDate, label, overdue := NextDueForTenant(tenant, time.Date(2026, time.July, 1, 0, 0, 0, 0, time.Local))
	want := time.Date(2026, time.August, 5, 0, 0, 0, 0, time.Local)
	if !dueDate.Equal(want) || label != "2026-08-05" || overdue {
		t.Fatalf("next due = %v/%q/%v, want %v/2026-08-05/false", dueDate, label, overdue, want)
	}

	dailyTenant := model.Tenant{RentType: model.RentTypeDaily}
	_, dailyLabel, dailyOverdue := NextDueForTenant(dailyTenant, time.Now())
	if dailyLabel != "每日" || !dailyOverdue {
		t.Fatalf("daily due = %q/%v, want 每日/true", dailyLabel, dailyOverdue)
	}
}
