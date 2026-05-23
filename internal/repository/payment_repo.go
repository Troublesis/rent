package repository

import (
	"strings"
	"time"

	"github.com/troublesis/rent/internal/model"
	"gorm.io/gorm"
)

type PaymentFilter struct {
	Paid         *bool
	Type         string
	TenantID     uint
	TenantStatus string
	Query        string
	Excluded     *bool
	Period       string
	Overdue      *bool
	Page         int
	Limit        int
	SortBy       string
	SortDir      string
	FromDate     time.Time
	ToDate       time.Time
}

type PaymentListResult struct {
	Payments []model.Payment
	Total    int64
}

type PaymentSummary struct {
	TotalUnpaidAmount    int
	TotalPaidAmount      int
	CheckoutPendingCount int64
	ExcludedCount        int64
}

type MonthlyIncomeRow struct {
	Month int
	Total int
}

type MonthlyIncomeRangeRow struct {
	Year  int
	Month int
	Total int
}

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) WithDB(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) ListPayments(filter PaymentFilter) ([]model.Payment, error) {
	joined := paymentFilterNeedsJoins(filter)
	query := r.paymentPreloadQuery()
	if joined {
		query = joinPaymentTenantRoom(query)
	}
	query = applyPaymentFilter(query, filter, time.Now())
	query = query.Order("payments.pay_date DESC, payments.created_at DESC")
	var payments []model.Payment
	if err := query.Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

func (r *PaymentRepository) ListPaymentsPage(filter PaymentFilter, now time.Time) (PaymentListResult, error) {
	joined := paymentFilterNeedsJoins(filter)
	countQuery := r.db.Model(&model.Payment{})
	if joined {
		countQuery = joinPaymentTenantRoom(countQuery)
	}
	countQuery = applyPaymentFilter(countQuery, filter, now)
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return PaymentListResult{}, err
	}

	page := normalizedPage(filter.Page)
	limit := normalizedLimit(filter.Limit)
	query := r.paymentPreloadQuery()
	if joined {
		query = joinPaymentTenantRoom(query)
	}
	query = applyPaymentFilter(query, filter, now)
	query = applyPaymentSort(query, filter, joined)
	query = query.Limit(limit).Offset((page - 1) * limit)

	var payments []model.Payment
	if err := query.Find(&payments).Error; err != nil {
		return PaymentListResult{}, err
	}
	return PaymentListResult{Payments: payments, Total: total}, nil
}

func (r *PaymentRepository) SummarizePayments(filter PaymentFilter, now time.Time) (PaymentSummary, error) {
	summaryFilter := filter
	summaryFilter.Paid = nil
	summaryFilter.Excluded = nil
	summaryFilter.Overdue = nil
	summaryQuery := func() *gorm.DB {
		query := joinPaymentTenantRoom(r.db.Model(&model.Payment{}))
		return applyPaymentFilter(query, summaryFilter, now)
	}

	var summary PaymentSummary
	if err := summaryQuery().Select("COALESCE(SUM(payments.amount), 0)").Where("payments.paid = ? AND payments.excluded = ?", false, false).Scan(&summary.TotalUnpaidAmount).Error; err != nil {
		return PaymentSummary{}, err
	}
	if err := summaryQuery().Select("COALESCE(SUM(payments.amount), 0)").Where("payments.paid = ? AND payments.excluded = ?", true, false).Scan(&summary.TotalPaidAmount).Error; err != nil {
		return PaymentSummary{}, err
	}
	if err := summaryQuery().Where("tenants.status = ? AND payments.excluded = ?", model.TenantStatusCheckout, false).Count(&summary.CheckoutPendingCount).Error; err != nil {
		return PaymentSummary{}, err
	}
	if err := summaryQuery().Where("payments.excluded = ?", true).Count(&summary.ExcludedCount).Error; err != nil {
		return PaymentSummary{}, err
	}
	return summary, nil
}

func (r *PaymentRepository) GetPayment(id uint) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.Preload("Tenant").Preload("Tenant.Room").Preload("Tenant.Payments", func(db *gorm.DB) *gorm.DB {
		return db.Order("pay_date ASC, created_at ASC")
	}).First(&payment, id).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepository) CreatePayment(payment *model.Payment) error {
	return r.db.Create(payment).Error
}

func (r *PaymentRepository) CreatePayments(payments []model.Payment) error {
	if len(payments) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&payments).Error
	})
}

func (r *PaymentRepository) UpdatePayment(payment *model.Payment) error {
	return r.db.Save(payment).Error
}

func (r *PaymentRepository) SumPaidByMonth(year int, month time.Month) (int, error) {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)
	var total int
	if err := r.db.Model(&model.Payment{}).Select("COALESCE(SUM(amount), 0)").Where("paid = ? AND excluded = ? AND pay_date >= ? AND pay_date < ?", true, false, start, end).Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *PaymentRepository) SumUnpaid() (int, error) {
	var total int
	if err := r.db.Model(&model.Payment{}).Select("COALESCE(SUM(amount), 0)").Where("paid = ? AND excluded = ?", false, false).Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *PaymentRepository) MonthlyIncome(year int) ([]MonthlyIncomeRow, error) {
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)
	rangeRows, err := r.MonthlyIncomeRange(start, start.AddDate(1, 0, 0))
	if err != nil {
		return nil, err
	}
	rows := make([]MonthlyIncomeRow, 0, len(rangeRows))
	for _, row := range rangeRows {
		rows = append(rows, MonthlyIncomeRow{Month: row.Month, Total: row.Total})
	}
	return rows, nil
}

func (r *PaymentRepository) MonthlyIncomeRange(start time.Time, end time.Time) ([]MonthlyIncomeRangeRow, error) {
	rows := make([]MonthlyIncomeRangeRow, 0)
	if err := r.db.Model(&model.Payment{}).
		Select("CAST(strftime('%Y', pay_date) AS INTEGER) AS year, CAST(strftime('%m', pay_date) AS INTEGER) AS month, COALESCE(SUM(amount), 0) AS total").
		Where("paid = ? AND excluded = ? AND pay_date >= ? AND pay_date < ?", true, false, start, end).
		Group("year, month").
		Order("year ASC, month ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *PaymentRepository) paymentPreloadQuery() *gorm.DB {
	return r.db.Model(&model.Payment{}).
		Preload("Tenant").
		Preload("Tenant.Room").
		Preload("Tenant.Payments", func(db *gorm.DB) *gorm.DB {
			return db.Order("pay_date ASC, created_at ASC")
		})
}

func applyPaymentFilter(query *gorm.DB, filter PaymentFilter, now time.Time) *gorm.DB {
	if filter.Paid != nil {
		query = query.Where("payments.paid = ?", *filter.Paid)
	}
	if filter.Type != "" {
		query = query.Where("payments.type = ?", filter.Type)
	}
	if filter.TenantID > 0 {
		query = query.Where("payments.tenant_id = ?", filter.TenantID)
	}
	if filter.Excluded != nil {
		query = query.Where("payments.excluded = ?", *filter.Excluded)
	}
	if strings.TrimSpace(filter.Query) != "" {
		like := "%" + strings.TrimSpace(filter.Query) + "%"
		query = query.Where("tenants.name LIKE ? OR tenants.phone LIKE ? OR rooms.room_no LIKE ?", like, like, like)
	}
	if status := strings.TrimSpace(filter.TenantStatus); status != "" {
		query = query.Where("tenants.status = ?", status)
	}
	if start, end, ok := paymentPeriodRange(filter.Period, now); ok {
		query = query.Where("payments.pay_date >= ? AND payments.pay_date < ?", start, end)
	}
	if !filter.FromDate.IsZero() {
		query = query.Where("payments.pay_date >= ?", dateOnly(filter.FromDate))
	}
	if !filter.ToDate.IsZero() {
		query = query.Where("payments.pay_date < ?", dateOnly(filter.ToDate).AddDate(0, 0, 1))
	}
	if filter.Overdue != nil && *filter.Overdue {
		today := dateOnly(now)
		query = query.Where("payments.paid = ? AND payments.excluded = ? AND payments.pay_date < ?", false, false, today)
	}
	return query
}

func applyPaymentSort(query *gorm.DB, filter PaymentFilter, joined bool) *gorm.DB {
	dir := normalizedSortDir(filter.SortDir)
	switch strings.TrimSpace(filter.SortBy) {
	case "tenant":
		if !joined {
			query = joinPaymentTenantRoom(query)
		}
		return query.Order("tenants.name " + dir).Order("rooms.room_no " + dir).Order("payments.pay_date ASC")
	case "amount":
		return query.Order("payments.amount " + dir).Order("payments.pay_date ASC")
	case "checkin_date":
		if !joined {
			query = joinPaymentTenantRoom(query)
		}
		return query.Order("tenants.checkin_date " + dir).Order("payments.pay_date ASC")
	case "status":
		return query.Order("payments.paid " + dir).Order("payments.pay_date ASC")
	case "type":
		return query.Order("payments.type " + dir).Order("payments.pay_date ASC")
	case "created_at":
		return query.Order("payments.created_at " + dir).Order("payments.pay_date ASC")
	case "next_pay_date", "pay_date", "":
		return query.Order("payments.pay_date " + dir).Order("payments.created_at ASC")
	default:
		return query.Order("payments.pay_date ASC").Order("payments.created_at ASC")
	}
}

func paymentFilterNeedsJoins(filter PaymentFilter) bool {
	return strings.TrimSpace(filter.Query) != "" || strings.TrimSpace(filter.TenantStatus) != ""
}

func joinPaymentTenantRoom(query *gorm.DB) *gorm.DB {
	return query.Joins("JOIN tenants ON tenants.id = payments.tenant_id").Joins("JOIN rooms ON rooms.id = tenants.room_id")
}

func paymentPeriodRange(period string, now time.Time) (time.Time, time.Time, bool) {
	today := dateOnly(now)
	switch period {
	case "today":
		return today, today.AddDate(0, 0, 1), true
	case "week":
		daysSinceMonday := (int(today.Weekday()) + 6) % 7
		start := today.AddDate(0, 0, -daysSinceMonday)
		return start, start.AddDate(0, 0, 7), true
	case "month", "current_month":
		start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
		return start, start.AddDate(0, 1, 0), true
	default:
		return time.Time{}, time.Time{}, false
	}
}

func normalizedPage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizedLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func normalizedSortDir(dir string) string {
	if strings.EqualFold(dir, "desc") {
		return "DESC"
	}
	return "ASC"
}

func dateOnly(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}
