package service

import (
	"fmt"
	"time"

	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
)

type PaymentInput struct {
	TenantID   uint
	AmountYuan string
	Type       string
	Paid       bool
	PayDate    time.Time
	Note       string
}

type PaymentService struct {
	paymentRepo *repository.PaymentRepository
	tenantRepo  *repository.TenantRepository
}

func NewPaymentService(paymentRepo *repository.PaymentRepository, tenantRepo *repository.TenantRepository) *PaymentService {
	return &PaymentService{paymentRepo: paymentRepo, tenantRepo: tenantRepo}
}

func (s *PaymentService) ListPayments(filter repository.PaymentFilter) ([]model.Payment, error) {
	return s.paymentRepo.ListPayments(filter)
}

func (s *PaymentService) RecordPayment(input PaymentInput) (*model.Payment, error) {
	payment, err := buildPayment(input)
	if err != nil {
		return nil, err
	}
	if _, err := s.tenantRepo.GetTenant(input.TenantID); err != nil {
		return nil, err
	}
	if err := s.paymentRepo.CreatePayment(payment); err != nil {
		return nil, err
	}
	return payment, nil
}

func (s *PaymentService) TogglePaid(id uint) error {
	payment, err := s.paymentRepo.GetPayment(id)
	if err != nil {
		return err
	}
	updatedPayment := *payment
	updatedPayment.Paid = !payment.Paid
	return s.paymentRepo.UpdatePayment(&updatedPayment)
}

func (s *PaymentService) SumPaidByMonth(year int, month time.Month) (int, error) {
	return s.paymentRepo.SumPaidByMonth(year, month)
}

func (s *PaymentService) SumUnpaid() (int, error) {
	return s.paymentRepo.SumUnpaid()
}

func (s *PaymentService) MonthlyIncome(year int) ([]repository.MonthlyIncomeRow, error) {
	return s.paymentRepo.MonthlyIncome(year)
}

func buildPayment(input PaymentInput) (*model.Payment, error) {
	if input.TenantID == 0 {
		return nil, fmt.Errorf("tenant is required")
	}
	amount, err := ParseYuanToFen(input.AmountYuan)
	if err != nil {
		return nil, fmt.Errorf("invalid payment amount: %w", err)
	}
	if amount <= 0 {
		return nil, fmt.Errorf("payment amount must be positive")
	}
	if !model.ValidPaymentType(input.Type) {
		return nil, fmt.Errorf("invalid payment type")
	}
	payDate := input.PayDate
	if payDate.IsZero() {
		payDate = time.Now()
	}
	return &model.Payment{
		TenantID: input.TenantID,
		Amount:   amount,
		Type:     input.Type,
		Paid:     input.Paid,
		PayDate:  payDate,
		Note:     input.Note,
	}, nil
}
