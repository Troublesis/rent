package service

import (
	"sort"
	"time"

	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
)

type DashboardSummary struct {
	TotalRooms                 int64
	VacantRooms                int64
	OccupiedRooms              int64
	ActiveTenants              int64
	OverdueCheckoutTenants     int64
	CurrentMonthIncome         int
	UnpaidAmount               int
	CurrentMonthReceivable     int
	NextSixMonthsReceivable    int
	NextTwelveMonthsReceivable int
}

type ProjectionSummary struct {
	Period    string
	Total     int
	Collected int
	Unpaid    int
	NotDue    int
	Items     []ProjectionItem
	Months    []ProjectionMonth
}

type ProjectionItem struct {
	TenantID     uint
	TenantName   string
	RoomNo       string
	DueDate      time.Time
	Amount       int
	Status       string
	RentType     string
	PaymentTerms string
}

type ProjectionMonth struct {
	Month       string
	Total       int
	TenantCount int
}

type DashboardService struct {
	roomRepo    *repository.RoomRepository
	tenantRepo  *repository.TenantRepository
	paymentRepo *repository.PaymentRepository
}

func NewDashboardService(roomRepo *repository.RoomRepository, tenantRepo *repository.TenantRepository, paymentRepo *repository.PaymentRepository) *DashboardService {
	return &DashboardService{roomRepo: roomRepo, tenantRepo: tenantRepo, paymentRepo: paymentRepo}
}

func (s *DashboardService) Summary(now time.Time) (DashboardSummary, error) {
	totalRooms, err := s.roomRepo.CountByStatus("")
	if err != nil {
		return DashboardSummary{}, err
	}
	vacantRooms, err := s.roomRepo.CountByStatus(model.RoomStatusVacant)
	if err != nil {
		return DashboardSummary{}, err
	}
	occupiedRooms, err := s.roomRepo.CountByStatus(model.RoomStatusOccupied)
	if err != nil {
		return DashboardSummary{}, err
	}
	tenants, err := s.tenantRepo.ListActiveTenantsWithPayments()
	if err != nil {
		return DashboardSummary{}, err
	}
	currentMonthIncome, err := s.paymentRepo.SumPaidByMonth(now.Year(), now.Month())
	if err != nil {
		return DashboardSummary{}, err
	}
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return DashboardSummary{
		TotalRooms:                 totalRooms,
		VacantRooms:                vacantRooms,
		OccupiedRooms:              occupiedRooms,
		ActiveTenants:              int64(len(tenants)),
		OverdueCheckoutTenants:     countLeaseExpiredTenants(tenants, now),
		CurrentMonthIncome:         currentMonthIncome,
		UnpaidAmount:               outstandingReceivable(tenants, now),
		CurrentMonthReceivable:     projectedReceivable(tenants, monthStart, monthStart.AddDate(0, 1, 0)),
		NextSixMonthsReceivable:    projectedReceivable(tenants, monthStart, monthStart.AddDate(0, 6, 0)),
		NextTwelveMonthsReceivable: projectedReceivable(tenants, monthStart, monthStart.AddDate(0, 12, 0)),
	}, nil
}

func (s *DashboardService) Projection(period string, now time.Time) (ProjectionSummary, error) {
	tenants, err := s.tenantRepo.ListActiveTenantsWithPayments()
	if err != nil {
		return ProjectionSummary{}, err
	}
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	months := projectionMonthCount(period)
	end := monthStart.AddDate(0, months, 0)
	items := projectedReceivableItems(tenants, monthStart, end, now)
	return buildProjectionSummary(period, items), nil
}

func countLeaseExpiredTenants(tenants []model.Tenant, now time.Time) int64 {
	total := int64(0)
	today := dateOnly(now)
	for _, tenant := range tenants {
		if tenant.LeaseEndDate == nil || tenant.Status != model.TenantStatusActive {
			continue
		}
		if dateOnly(*tenant.LeaseEndDate).Before(today) {
			total++
		}
	}
	return total
}

func projectionMonthCount(period string) int {
	switch period {
	case "6months":
		return 6
	case "12months":
		return 12
	default:
		return 1
	}
}

func projectedReceivableItems(tenants []model.Tenant, start time.Time, end time.Time, now time.Time) []ProjectionItem {
	items := make([]ProjectionItem, 0)
	for _, tenant := range tenants {
		for _, due := range RentDueSchedule(tenant, end.AddDate(0, 0, -1)) {
			if due.DueDate.Before(dateOnly(start)) || !due.DueDate.Before(dateOnly(end)) {
				continue
			}
			items = append(items, ProjectionItem{
				TenantID:     tenant.ID,
				TenantName:   tenant.Name,
				RoomNo:       tenant.Room.RoomNo,
				DueDate:      due.DueDate,
				Amount:       due.Amount,
				Status:       projectionPaymentStatus(tenant.Payments, due.DueDate, now),
				RentType:     tenant.RentType,
				PaymentTerms: tenant.PaymentTerms,
			})
		}
	}
	return items
}

func projectionPaymentStatus(payments []model.Payment, dueDate time.Time, now time.Time) string {
	for _, payment := range payments {
		if payment.Type == model.PaymentTypeRent && !payment.Excluded && payment.Paid && sameCalendarDate(payment.PayDate, dueDate) {
			return "已收"
		}
	}
	if dateOnly(dueDate).After(dateOnly(now)) {
		return "未到期"
	}
	return "未收"
}

func buildProjectionSummary(period string, items []ProjectionItem) ProjectionSummary {
	summary := ProjectionSummary{Period: period, Items: items, Months: projectionMonths(items)}
	for _, item := range items {
		summary.Total += item.Amount
		switch item.Status {
		case "已收":
			summary.Collected += item.Amount
		case "未到期":
			summary.NotDue += item.Amount
		default:
			summary.Unpaid += item.Amount
		}
	}
	return summary
}

func projectionMonths(items []ProjectionItem) []ProjectionMonth {
	monthTotals := map[string]int{}
	tenantMonths := map[string]map[uint]bool{}
	order := make([]string, 0)
	for _, item := range items {
		month := item.DueDate.Format("2006-01")
		if _, ok := monthTotals[month]; !ok {
			order = append(order, month)
			tenantMonths[month] = map[uint]bool{}
		}
		monthTotals[month] += item.Amount
		tenantMonths[month][item.TenantID] = true
	}
	sort.Strings(order)
	months := make([]ProjectionMonth, 0, len(order))
	for _, month := range order {
		months = append(months, ProjectionMonth{Month: month, Total: monthTotals[month], TenantCount: len(tenantMonths[month])})
	}
	return months
}

func outstandingReceivable(tenants []model.Tenant, now time.Time) int {
	total := 0
	for _, tenant := range tenants {
		expected := expectedReceivableUntil(tenant, now)
		paid := paidRentUntil(tenant, now)
		excluded := excludedUnpaidRentUntil(tenant, now)
		outstanding := expected - paid - excluded
		if outstanding > 0 {
			total += outstanding
		}
	}
	return total
}

func expectedReceivableUntil(tenant model.Tenant, now time.Time) int {
	end := dateOnly(now).AddDate(0, 0, 1)
	return projectedReceivableForTenant(tenant, dateOnly(tenant.CheckinDate), end)
}

func projectedReceivable(tenants []model.Tenant, start time.Time, end time.Time) int {
	total := 0
	for _, tenant := range tenants {
		total += projectedReceivableForTenant(tenant, start, end)
	}
	return total
}

func projectedReceivableForTenant(tenant model.Tenant, start time.Time, end time.Time) int {
	if tenant.RentPrice <= 0 || !end.After(start) {
		return 0
	}
	checkinDate := dateOnly(tenant.CheckinDate)
	windowStart := dateOnly(start)
	windowEnd := dateOnly(end)
	if checkinDate.After(windowStart) {
		windowStart = checkinDate
	}
	if !windowEnd.After(windowStart) {
		return 0
	}
	if model.RentTypeOrDefault(tenant.RentType) == model.RentTypeDaily {
		return daysBetween(windowStart, windowEnd) * tenant.RentPrice
	}
	cycleMonths := model.PaymentTermsMonths(tenant.PaymentTerms)
	dueDate := checkinDate
	for dueDate.Before(windowStart) {
		dueDate = dueDate.AddDate(0, cycleMonths, 0)
	}
	total := 0
	for dueDate.Before(windowEnd) {
		total += tenant.RentPrice * cycleMonths
		dueDate = dueDate.AddDate(0, cycleMonths, 0)
	}
	return total
}

func paidRentUntil(tenant model.Tenant, now time.Time) int {
	total := 0
	end := dateOnly(now).AddDate(0, 0, 1)
	checkinDate := dateOnly(tenant.CheckinDate)
	for _, payment := range tenant.Payments {
		if payment.Type != model.PaymentTypeRent || !payment.Paid || payment.Excluded {
			continue
		}
		payDate := dateOnly(payment.PayDate)
		if !payDate.Before(checkinDate) && payDate.Before(end) {
			total += payment.Amount
		}
	}
	return total
}

func excludedUnpaidRentUntil(tenant model.Tenant, now time.Time) int {
	total := 0
	end := dateOnly(now).AddDate(0, 0, 1)
	checkinDate := dateOnly(tenant.CheckinDate)
	for _, payment := range tenant.Payments {
		if payment.Type != model.PaymentTypeRent || payment.Paid || !payment.Excluded {
			continue
		}
		payDate := dateOnly(payment.PayDate)
		if !payDate.Before(checkinDate) && payDate.Before(end) {
			total += payment.Amount
		}
	}
	return total
}

func dateOnly(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func daysBetween(start time.Time, end time.Time) int {
	if !end.After(start) {
		return 0
	}
	return int(end.Sub(start).Hours() / 24)
}
