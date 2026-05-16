package service

import (
	"time"

	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
)

type DashboardSummary struct {
	TotalRooms         int64
	VacantRooms        int64
	OccupiedRooms      int64
	ActiveTenants      int64
	CurrentMonthIncome int
	UnpaidAmount       int
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
	activeTenants, err := s.tenantRepo.CountActiveTenants()
	if err != nil {
		return DashboardSummary{}, err
	}
	currentMonthIncome, err := s.paymentRepo.SumPaidByMonth(now.Year(), now.Month())
	if err != nil {
		return DashboardSummary{}, err
	}
	unpaidAmount, err := s.paymentRepo.SumUnpaid()
	if err != nil {
		return DashboardSummary{}, err
	}
	return DashboardSummary{
		TotalRooms:         totalRooms,
		VacantRooms:        vacantRooms,
		OccupiedRooms:      occupiedRooms,
		ActiveTenants:      activeTenants,
		CurrentMonthIncome: currentMonthIncome,
		UnpaidAmount:       unpaidAmount,
	}, nil
}
