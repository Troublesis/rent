package repository

import (
	"testing"
	"time"

	"github.com/troublesis/rent/internal/model"
	"gorm.io/gorm"
)

func TestPaymentRepositoryListPaymentsPageSortsAndPaginates(t *testing.T) {
	db := newTestDB(t)
	repo := NewPaymentRepository(db)
	tenant := createPaymentRepoTenant(t, db, "P101", "排序租客", "13800001001")

	createPaymentRepoPayment(t, db, tenant.ID, 300000, time.Date(2026, time.May, 20, 0, 0, 0, 0, time.Local), false, false)
	createPaymentRepoPayment(t, db, tenant.ID, 100000, time.Date(2026, time.May, 18, 0, 0, 0, 0, time.Local), false, false)
	createPaymentRepoPayment(t, db, tenant.ID, 200000, time.Date(2026, time.May, 19, 0, 0, 0, 0, time.Local), false, false)

	result, err := repo.ListPaymentsPage(PaymentFilter{Page: 1, Limit: 2, SortBy: "next_pay_date", SortDir: "asc"}, time.Date(2026, time.May, 17, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("ListPaymentsPage returned error: %v", err)
	}
	if result.Total != 3 {
		t.Fatalf("Total = %d, want 3", result.Total)
	}
	if len(result.Payments) != 2 {
		t.Fatalf("len(Payments) = %d, want 2", len(result.Payments))
	}
	if result.Payments[0].Amount != 100000 || result.Payments[1].Amount != 200000 {
		t.Fatalf("payments order = %#v, want amount 100000 then 200000", result.Payments)
	}
}

func TestPaymentRepositoryListPaymentsPageAppliesPeriodAndOverdue(t *testing.T) {
	db := newTestDB(t)
	repo := NewPaymentRepository(db)
	tenant := createPaymentRepoTenant(t, db, "P102", "周期租客", "13800001002")
	now := time.Date(2026, time.May, 17, 12, 0, 0, 0, time.Local)

	createPaymentRepoPayment(t, db, tenant.ID, 100000, time.Date(2026, time.May, 17, 0, 0, 0, 0, time.Local), false, false)
	createPaymentRepoPayment(t, db, tenant.ID, 200000, time.Date(2026, time.May, 13, 0, 0, 0, 0, time.Local), true, false)
	createPaymentRepoPayment(t, db, tenant.ID, 300000, time.Date(2026, time.May, 1, 0, 0, 0, 0, time.Local), false, false)
	createPaymentRepoPayment(t, db, tenant.ID, 400000, time.Date(2026, time.April, 30, 0, 0, 0, 0, time.Local), false, true)

	todayResult, err := repo.ListPaymentsPage(PaymentFilter{Period: "today", Page: 1, Limit: 20}, now)
	if err != nil {
		t.Fatalf("today ListPaymentsPage returned error: %v", err)
	}
	if todayResult.Total != 1 || len(todayResult.Payments) != 1 || todayResult.Payments[0].Amount != 100000 {
		t.Fatalf("todayResult = %#v, want only today payment", todayResult)
	}

	weekResult, err := repo.ListPaymentsPage(PaymentFilter{Period: "week", Page: 1, Limit: 20}, now)
	if err != nil {
		t.Fatalf("week ListPaymentsPage returned error: %v", err)
	}
	if weekResult.Total != 2 {
		t.Fatalf("week Total = %d, want 2", weekResult.Total)
	}

	overdue := true
	overdueResult, err := repo.ListPaymentsPage(PaymentFilter{Overdue: &overdue, Page: 1, Limit: 20}, now)
	if err != nil {
		t.Fatalf("overdue ListPaymentsPage returned error: %v", err)
	}
	if overdueResult.Total != 1 || len(overdueResult.Payments) != 1 || overdueResult.Payments[0].Amount != 300000 {
		t.Fatalf("overdueResult = %#v, want only unpaid non-excluded overdue payment", overdueResult)
	}
}

func TestPaymentRepositorySummarizePayments(t *testing.T) {
	db := newTestDB(t)
	repo := NewPaymentRepository(db)
	activeTenant := createPaymentRepoTenantWithStatus(t, db, "P103", "在租汇总", "13800001003", model.TenantStatusActive)
	checkoutTenant := createPaymentRepoTenantWithStatus(t, db, "P104", "退租汇总", "13800001004", model.TenantStatusCheckout)
	now := time.Date(2026, time.May, 17, 12, 0, 0, 0, time.Local)

	createPaymentRepoPayment(t, db, activeTenant.ID, 100000, time.Date(2026, time.May, 1, 0, 0, 0, 0, time.Local), false, false)
	createPaymentRepoPayment(t, db, activeTenant.ID, 200000, time.Date(2026, time.May, 2, 0, 0, 0, 0, time.Local), true, false)
	createPaymentRepoPayment(t, db, checkoutTenant.ID, 300000, time.Date(2026, time.May, 3, 0, 0, 0, 0, time.Local), false, false)
	createPaymentRepoPayment(t, db, checkoutTenant.ID, 400000, time.Date(2026, time.May, 4, 0, 0, 0, 0, time.Local), true, false)
	createPaymentRepoPayment(t, db, checkoutTenant.ID, 500000, time.Date(2026, time.May, 5, 0, 0, 0, 0, time.Local), false, true)

	summary, err := repo.SummarizePayments(PaymentFilter{Period: "month"}, now)
	if err != nil {
		t.Fatalf("SummarizePayments returned error: %v", err)
	}
	if summary.TotalUnpaidAmount != 400000 {
		t.Fatalf("TotalUnpaidAmount = %d, want 400000", summary.TotalUnpaidAmount)
	}
	if summary.TotalPaidAmount != 600000 {
		t.Fatalf("TotalPaidAmount = %d, want 600000", summary.TotalPaidAmount)
	}
	if summary.CheckoutPendingCount != 1 {
		t.Fatalf("CheckoutPendingCount = %d, want 1", summary.CheckoutPendingCount)
	}
	if summary.ExcludedCount != 1 {
		t.Fatalf("ExcludedCount = %d, want 1", summary.ExcludedCount)
	}
}

func TestPaymentRepositoryMonthlyIncomeRange(t *testing.T) {
	db := newTestDB(t)
	repo := NewPaymentRepository(db)
	tenant := createPaymentRepoTenant(t, db, "P105", "跨年租客", "13800001005")

	createPaymentRepoPayment(t, db, tenant.ID, 100000, time.Date(2025, time.December, 15, 0, 0, 0, 0, time.Local), true, false)
	createPaymentRepoPayment(t, db, tenant.ID, 200000, time.Date(2026, time.January, 15, 0, 0, 0, 0, time.Local), true, false)
	createPaymentRepoPayment(t, db, tenant.ID, 300000, time.Date(2026, time.January, 16, 0, 0, 0, 0, time.Local), false, false)
	createPaymentRepoPayment(t, db, tenant.ID, 400000, time.Date(2026, time.January, 17, 0, 0, 0, 0, time.Local), true, true)
	createPaymentRepoPayment(t, db, tenant.ID, 500000, time.Date(2026, time.February, 1, 0, 0, 0, 0, time.Local), true, false)

	rows, err := repo.MonthlyIncomeRange(time.Date(2025, time.December, 1, 0, 0, 0, 0, time.Local), time.Date(2026, time.February, 1, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("MonthlyIncomeRange returned error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2", len(rows))
	}
	if rows[0].Year != 2025 || rows[0].Month != 12 || rows[0].Total != 100000 {
		t.Fatalf("rows[0] = %#v, want 2025-12 total 100000", rows[0])
	}
	if rows[1].Year != 2026 || rows[1].Month != 1 || rows[1].Total != 200000 {
		t.Fatalf("rows[1] = %#v, want 2026-01 total 200000", rows[1])
	}
}

func createPaymentRepoTenant(t *testing.T, db interface {
	Create(value interface{}) *gorm.DB
}, roomNo string, name string, phone string) model.Tenant {
	t.Helper()
	return createPaymentRepoTenantWithStatus(t, db, roomNo, name, phone, model.TenantStatusActive)
}

func createPaymentRepoTenantWithStatus(t *testing.T, db interface {
	Create(value interface{}) *gorm.DB
}, roomNo string, name string, phone string, status string) model.Tenant {
	t.Helper()
	room := model.Room{RoomNo: roomNo, Title: roomNo + " 房源", RentType: model.RentTypeMonthly, RentPrice: 100000, PaymentTerms: model.PaymentTerms1M1D, Deposit: 100000, Status: model.RoomStatusOccupied}
	if err := db.Create(&room).Error; err != nil {
		t.Fatalf("create room: %v", err)
	}
	tenant := model.Tenant{Name: name, Phone: phone, RoomID: room.ID, CheckinDate: time.Date(2026, time.May, 1, 0, 0, 0, 0, time.Local), RentPrice: 100000, RentType: model.RentTypeMonthly, PaymentTerms: model.PaymentTerms1M1D, Status: status}
	if err := db.Create(&tenant).Error; err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	return tenant
}

func createPaymentRepoPayment(t *testing.T, db interface {
	Create(value interface{}) *gorm.DB
}, tenantID uint, amount int, payDate time.Time, paid bool, excluded bool) model.Payment {
	t.Helper()
	payment := model.Payment{TenantID: tenantID, Amount: amount, Type: model.PaymentTypeRent, Paid: paid, PayDate: payDate, Excluded: excluded}
	if err := db.Create(&payment).Error; err != nil {
		t.Fatalf("create payment: %v", err)
	}
	return payment
}
