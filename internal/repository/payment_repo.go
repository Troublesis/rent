package repository

import (
	"time"

	"github.com/troublesis/rent/internal/model"
	"gorm.io/gorm"
)

type PaymentFilter struct {
	Paid     *bool
	Type     string
	TenantID uint
}

type MonthlyIncomeRow struct {
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
	query := r.db.Model(&model.Payment{}).Preload("Tenant").Preload("Tenant.Room").Order("pay_date DESC, created_at DESC")
	if filter.Paid != nil {
		query = query.Where("paid = ?", *filter.Paid)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.TenantID > 0 {
		query = query.Where("tenant_id = ?", filter.TenantID)
	}
	var payments []model.Payment
	if err := query.Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

func (r *PaymentRepository) GetPayment(id uint) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.Preload("Tenant").First(&payment, id).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepository) CreatePayment(payment *model.Payment) error {
	return r.db.Create(payment).Error
}

func (r *PaymentRepository) UpdatePayment(payment *model.Payment) error {
	return r.db.Save(payment).Error
}

func (r *PaymentRepository) SumPaidByMonth(year int, month time.Month) (int, error) {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)
	var total int
	if err := r.db.Model(&model.Payment{}).Select("COALESCE(SUM(amount), 0)").Where("paid = ? AND pay_date >= ? AND pay_date < ?", true, start, end).Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *PaymentRepository) SumUnpaid() (int, error) {
	var total int
	if err := r.db.Model(&model.Payment{}).Select("COALESCE(SUM(amount), 0)").Where("paid = ?", false).Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *PaymentRepository) MonthlyIncome(year int) ([]MonthlyIncomeRow, error) {
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(1, 0, 0)
	rows := make([]MonthlyIncomeRow, 0, 12)
	if err := r.db.Model(&model.Payment{}).
		Select("CAST(strftime('%m', pay_date) AS INTEGER) AS month, COALESCE(SUM(amount), 0) AS total").
		Where("paid = ? AND pay_date >= ? AND pay_date < ?", true, start, end).
		Group("month").
		Order("month ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
