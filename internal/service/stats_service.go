package service

import (
	"fmt"
	"time"

	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
)

const maxStatsRangeYears = 5

type StatsFilter struct {
	Year  int
	Start time.Time
	End   time.Time
}

type StatsRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
	Year  int    `json:"year"`
}

type StatsOverview struct {
	Range                StatsRange `json:"range"`
	TotalPaidFen         int        `json:"total_paid_fen"`
	AverageMonthlyFen    int        `json:"average_monthly_fen"`
	PeakMonth            string     `json:"peak_month"`
	PeakMonthPaidFen     int        `json:"peak_month_paid_fen"`
	TotalRooms           int64      `json:"total_rooms"`
	OccupiedRooms        int64      `json:"occupied_rooms"`
	ActiveTenants        int64      `json:"active_tenants"`
	OccupancyRate        float64    `json:"occupancy_rate"`
	ApproximateOccupancy bool       `json:"approximate_occupancy"`
}

type MonthlyIncomeReport struct {
	Range   StatsRange           `json:"range"`
	Labels  []string             `json:"labels"`
	Totals  []int                `json:"totals_fen"`
	Months  []MonthlyIncomeItem  `json:"months"`
	Summary MonthlyIncomeSummary `json:"summary"`
}

type MonthlyIncomeItem struct {
	Month   string `json:"month"`
	Label   string `json:"label"`
	PaidFen int    `json:"paid_fen"`
}

type MonthlyIncomeSummary struct {
	TotalFen          int    `json:"total_fen"`
	AverageMonthlyFen int    `json:"average_monthly_fen"`
	PeakMonth         string `json:"peak_month"`
	PeakMonthPaidFen  int    `json:"peak_month_paid_fen"`
}

type MonthlyOccupancyReport struct {
	Range       StatsRange             `json:"range"`
	Labels      []string               `json:"labels"`
	Rates       []float64              `json:"rates"`
	Months      []MonthlyOccupancyItem `json:"months"`
	Approximate bool                   `json:"approximate"`
	Note        string                 `json:"note"`
}

type MonthlyOccupancyItem struct {
	Month         string  `json:"month"`
	Label         string  `json:"label"`
	OccupiedRooms int     `json:"occupied_rooms"`
	TotalRooms    int     `json:"total_rooms"`
	Rate          float64 `json:"rate"`
}

type StatsProjectionReport struct {
	Period       string            `json:"period"`
	TotalFen     int               `json:"total_fen"`
	CollectedFen int               `json:"collected_fen"`
	UnpaidFen    int               `json:"unpaid_fen"`
	NotDueFen    int               `json:"not_due_fen"`
	Months       []ProjectionMonth `json:"months"`
	Note         string            `json:"note"`
}

type StatsService struct {
	roomRepo         *repository.RoomRepository
	tenantRepo       *repository.TenantRepository
	paymentRepo      *repository.PaymentRepository
	dashboardService *DashboardService
}

func NewStatsService(roomRepo *repository.RoomRepository, tenantRepo *repository.TenantRepository, paymentRepo *repository.PaymentRepository, dashboardService *DashboardService) *StatsService {
	return &StatsService{roomRepo: roomRepo, tenantRepo: tenantRepo, paymentRepo: paymentRepo, dashboardService: dashboardService}
}

func NewYearStatsFilter(year int, now time.Time) (StatsFilter, error) {
	if year == 0 {
		year = now.Year()
	}
	if year < 2000 || year > 2100 {
		return StatsFilter{}, fmt.Errorf("年份不正确")
	}
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, now.Location())
	return StatsFilter{Year: year, Start: start, End: start.AddDate(1, 0, 0)}, nil
}

func NewDateRangeStatsFilter(start time.Time, end time.Time, now time.Time) (StatsFilter, error) {
	start = dateOnly(start)
	end = dateOnly(end)
	if start.IsZero() || end.IsZero() || end.Before(start) {
		return StatsFilter{}, fmt.Errorf("时间范围不正确")
	}
	endExclusive := end.AddDate(0, 0, 1)
	if endExclusive.After(start.AddDate(maxStatsRangeYears, 0, 1)) {
		return StatsFilter{}, fmt.Errorf("时间范围不能超过 5 年")
	}
	return StatsFilter{Year: now.Year(), Start: start, End: endExclusive}, nil
}

func (s *StatsService) Overview(filter StatsFilter) (StatsOverview, error) {
	income, err := s.MonthlyIncome(filter)
	if err != nil {
		return StatsOverview{}, err
	}
	totalRooms, err := s.roomRepo.CountByStatus("")
	if err != nil {
		return StatsOverview{}, err
	}
	occupiedRooms, err := s.roomRepo.CountByStatus(model.RoomStatusOccupied)
	if err != nil {
		return StatsOverview{}, err
	}
	activeTenants, err := s.tenantRepo.CountActiveTenants()
	if err != nil {
		return StatsOverview{}, err
	}
	occupancyRate := 0.0
	if totalRooms > 0 {
		occupancyRate = float64(occupiedRooms) / float64(totalRooms)
	}
	return StatsOverview{
		Range:                filter.Range(),
		TotalPaidFen:         income.Summary.TotalFen,
		AverageMonthlyFen:    income.Summary.AverageMonthlyFen,
		PeakMonth:            income.Summary.PeakMonth,
		PeakMonthPaidFen:     income.Summary.PeakMonthPaidFen,
		TotalRooms:           totalRooms,
		OccupiedRooms:        occupiedRooms,
		ActiveTenants:        activeTenants,
		OccupancyRate:        occupancyRate,
		ApproximateOccupancy: true,
	}, nil
}

func (s *StatsService) MonthlyIncome(filter StatsFilter) (MonthlyIncomeReport, error) {
	rows, err := s.paymentRepo.MonthlyIncomeRange(filter.Start, filter.End)
	if err != nil {
		return MonthlyIncomeReport{}, err
	}
	totalsByMonth := map[string]int{}
	for _, row := range rows {
		key := fmt.Sprintf("%04d-%02d", row.Year, row.Month)
		totalsByMonth[key] = row.Total
	}
	periods := filter.Months()
	labels := make([]string, 0, len(periods))
	totals := make([]int, 0, len(periods))
	months := make([]MonthlyIncomeItem, 0, len(periods))
	summary := MonthlyIncomeSummary{}
	for _, period := range periods {
		key := period.Format("2006-01")
		label := monthLabel(period)
		total := totalsByMonth[key]
		labels = append(labels, label)
		totals = append(totals, total)
		months = append(months, MonthlyIncomeItem{Month: key, Label: label, PaidFen: total})
		summary.TotalFen += total
		if total > summary.PeakMonthPaidFen || summary.PeakMonth == "" {
			summary.PeakMonth = key
			summary.PeakMonthPaidFen = total
		}
	}
	if len(periods) > 0 {
		summary.AverageMonthlyFen = summary.TotalFen / len(periods)
	}
	return MonthlyIncomeReport{Range: filter.Range(), Labels: labels, Totals: totals, Months: months, Summary: summary}, nil
}

func (s *StatsService) MonthlyOccupancy(filter StatsFilter) (MonthlyOccupancyReport, error) {
	rooms, err := s.roomRepo.ListRooms(repository.RoomFilter{})
	if err != nil {
		return MonthlyOccupancyReport{}, err
	}
	tenants, err := s.tenantRepo.ListTenantsOverlappingPeriod(filter.Start, filter.End)
	if err != nil {
		return MonthlyOccupancyReport{}, err
	}
	periods := filter.Months()
	labels := make([]string, 0, len(periods))
	rates := make([]float64, 0, len(periods))
	months := make([]MonthlyOccupancyItem, 0, len(periods))
	totalRooms := len(rooms)
	for _, period := range periods {
		periodStart := period
		periodEnd := period.AddDate(0, 1, 0)
		occupiedRoomIDs := map[uint]bool{}
		for _, tenant := range tenants {
			if tenantOverlapsPeriod(tenant, periodStart, periodEnd) {
				occupiedRoomIDs[tenant.RoomID] = true
			}
		}
		rate := 0.0
		if totalRooms > 0 {
			rate = float64(len(occupiedRoomIDs)) / float64(totalRooms)
		}
		label := monthLabel(period)
		labels = append(labels, label)
		rates = append(rates, rate)
		months = append(months, MonthlyOccupancyItem{Month: period.Format("2006-01"), Label: label, OccupiedRooms: len(occupiedRoomIDs), TotalRooms: totalRooms, Rate: rate})
	}
	return MonthlyOccupancyReport{
		Range:       filter.Range(),
		Labels:      labels,
		Rates:       rates,
		Months:      months,
		Approximate: true,
		Note:        "出租率根据租客入住和退租日期估算，未包含房态历史变更。",
	}, nil
}

func (s *StatsService) Projection(period string, now time.Time) (StatsProjectionReport, error) {
	projection, err := s.dashboardService.Projection(period, now)
	if err != nil {
		return StatsProjectionReport{}, err
	}
	return StatsProjectionReport{
		Period:       projection.Period,
		TotalFen:     projection.Total,
		CollectedFen: projection.Collected,
		UnpaidFen:    projection.Unpaid,
		NotDueFen:    projection.NotDue,
		Months:       projection.Months,
		Note:         "以下数据为预测值，基于当前在租租客和租金周期估算，不含未来新增租约。",
	}, nil
}

func (f StatsFilter) Range() StatsRange {
	return StatsRange{Start: f.Start.Format("2006-01-02"), End: f.End.AddDate(0, 0, -1).Format("2006-01-02"), Year: f.Year}
}

func (f StatsFilter) Months() []time.Time {
	months := make([]time.Time, 0)
	for current := monthStart(f.Start); current.Before(f.End); current = current.AddDate(0, 1, 0) {
		months = append(months, current)
	}
	return months
}

func monthStart(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, value.Location())
}

func monthLabel(value time.Time) string {
	return fmt.Sprintf("%d月", int(value.Month()))
}

func tenantOverlapsPeriod(tenant model.Tenant, start time.Time, end time.Time) bool {
	checkinDate := dateOnly(tenant.CheckinDate)
	if checkinDate.IsZero() || !checkinDate.Before(end) {
		return false
	}
	if tenant.CheckoutDate != nil && dateOnly(*tenant.CheckoutDate).Before(start) {
		return false
	}
	return true
}
