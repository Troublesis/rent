package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
	"gorm.io/gorm"
)

type TenantInput struct {
	Name             string
	Phone            string
	EmergencyContact string
	RoomID           uint
	CheckinDate      time.Time
	RentPriceYuan    string
	DepositYuan      string
}

type TenantService struct {
	db         *gorm.DB
	tenantRepo *repository.TenantRepository
	roomRepo   *repository.RoomRepository
}

func NewTenantService(db *gorm.DB, tenantRepo *repository.TenantRepository, roomRepo *repository.RoomRepository) *TenantService {
	return &TenantService{db: db, tenantRepo: tenantRepo, roomRepo: roomRepo}
}

func (s *TenantService) ListTenants(filter repository.TenantFilter) ([]model.Tenant, error) {
	return s.tenantRepo.ListTenants(filter)
}

func (s *TenantService) GetTenant(id uint) (*model.Tenant, error) {
	return s.tenantRepo.GetTenant(id)
}

func (s *TenantService) CheckInTenant(input TenantInput) (*model.Tenant, error) {
	tenant, err := buildTenant(input)
	if err != nil {
		return nil, err
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		roomRepo := s.roomRepo.WithDB(tx)
		tenantRepo := s.tenantRepo.WithDB(tx)
		room, err := roomRepo.GetRoom(input.RoomID)
		if err != nil {
			return err
		}
		if room.Status != model.RoomStatusVacant {
			return fmt.Errorf("room is not vacant")
		}
		if _, err := tenantRepo.GetActiveTenantByRoomID(input.RoomID); err == nil {
			return fmt.Errorf("room already has an active tenant")
		} else if !repository.IsNotFound(err) {
			return err
		}
		if err := tenantRepo.CreateTenant(tenant); err != nil {
			return err
		}
		updatedRoom := *room
		updatedRoom.Status = model.RoomStatusOccupied
		return roomRepo.UpdateRoom(&updatedRoom)
	}); err != nil {
		return nil, err
	}

	return tenant, nil
}

func (s *TenantService) CheckOutTenant(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		tenantRepo := s.tenantRepo.WithDB(tx)
		roomRepo := s.roomRepo.WithDB(tx)
		tenant, err := tenantRepo.GetTenant(id)
		if err != nil {
			return err
		}
		if tenant.Status != model.TenantStatusActive {
			return fmt.Errorf("tenant is not active")
		}
		now := time.Now()
		updatedTenant := *tenant
		updatedTenant.Status = model.TenantStatusCheckout
		updatedTenant.CheckoutDate = &now
		if err := tenantRepo.UpdateTenant(&updatedTenant); err != nil {
			return err
		}
		room, err := roomRepo.GetRoom(tenant.RoomID)
		if err != nil {
			return err
		}
		updatedRoom := *room
		updatedRoom.Status = model.RoomStatusVacant
		return roomRepo.UpdateRoom(&updatedRoom)
	})
}

func buildTenant(input TenantInput) (*model.Tenant, error) {
	name := strings.TrimSpace(input.Name)
	phone := strings.TrimSpace(input.Phone)
	if name == "" {
		return nil, fmt.Errorf("tenant name is required")
	}
	if phone == "" {
		return nil, fmt.Errorf("tenant phone is required")
	}
	if input.RoomID == 0 {
		return nil, fmt.Errorf("room is required")
	}
	rentPrice, err := ParseYuanToFen(input.RentPriceYuan)
	if err != nil {
		return nil, fmt.Errorf("invalid rent price: %w", err)
	}
	deposit, err := ParseYuanToFen(input.DepositYuan)
	if err != nil {
		return nil, fmt.Errorf("invalid deposit: %w", err)
	}
	checkinDate := input.CheckinDate
	if checkinDate.IsZero() {
		checkinDate = time.Now()
	}
	return &model.Tenant{
		Name:             name,
		Phone:            phone,
		EmergencyContact: strings.TrimSpace(input.EmergencyContact),
		RoomID:           input.RoomID,
		CheckinDate:      checkinDate,
		RentPrice:        rentPrice,
		Deposit:          deposit,
		Status:           model.TenantStatusActive,
	}, nil
}
