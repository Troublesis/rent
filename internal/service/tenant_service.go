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
	Gender           string
	RoomID           uint
	CheckinDate      time.Time
	LeaseEndDate     time.Time
	RentPriceYuan    string
	RentType         string
	PaymentTerms     string
	DepositYuan      string
	Notes            string
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

func (s *TenantService) ListTenantsByRoomID(roomID uint) ([]model.Tenant, error) {
	return s.tenantRepo.ListTenantsByRoomID(roomID)
}

func (s *TenantService) CheckInTenant(input TenantInput) (*model.Tenant, error) {
	var tenant *model.Tenant
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		roomRepo := s.roomRepo.WithDB(tx)
		tenantRepo := s.tenantRepo.WithDB(tx)
		room, err := roomRepo.GetRoom(input.RoomID)
		if err != nil {
			return err
		}
		if room.Status != model.RoomStatusVacant {
			return fmt.Errorf("房源不是空置状态")
		}
		if _, err := tenantRepo.GetActiveTenantByRoomID(input.RoomID); err == nil {
			return fmt.Errorf("该房源已有在租租客")
		} else if !repository.IsNotFound(err) {
			return err
		}
		resolvedInput := tenantInputWithRoomDefaults(input, *room)
		builtTenant, err := buildTenant(resolvedInput)
		if err != nil {
			return err
		}
		if err := tenantRepo.CreateTenant(builtTenant); err != nil {
			return err
		}
		updatedRoom := *room
		updatedRoom.Status = model.RoomStatusOccupied
		if err := roomRepo.UpdateRoom(&updatedRoom); err != nil {
			return err
		}
		tenant = builtTenant
		return nil
	}); err != nil {
		return nil, err
	}
	return tenant, nil
}

func (s *TenantService) UpdateTenant(id uint, input TenantInput) (*model.Tenant, error) {
	var tenant *model.Tenant
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		tenantRepo := s.tenantRepo.WithDB(tx)
		roomRepo := s.roomRepo.WithDB(tx)
		current, err := tenantRepo.GetTenant(id)
		if err != nil {
			return err
		}
		if current.Status != model.TenantStatusActive {
			return fmt.Errorf("只能编辑在租租客")
		}
		selectedRoom, err := roomRepo.GetRoom(input.RoomID)
		if err != nil {
			return err
		}
		if input.RoomID != current.RoomID {
			if selectedRoom.Status != model.RoomStatusVacant {
				return fmt.Errorf("目标房源不是空置状态")
			}
			if activeTenant, err := tenantRepo.GetActiveTenantByRoomID(input.RoomID); err == nil && activeTenant.ID != current.ID {
				return fmt.Errorf("目标房源已有在租租客")
			} else if err != nil && !repository.IsNotFound(err) {
				return err
			}
		}
		builtTenant, err := buildTenant(input)
		if err != nil {
			return err
		}
		updatedTenant := tenantWithEdits(*current, *builtTenant)
		if err := tenantRepo.UpdateTenant(&updatedTenant); err != nil {
			return err
		}
		if input.RoomID != current.RoomID {
			currentRoom, err := roomRepo.GetRoom(current.RoomID)
			if err != nil {
				return err
			}
			vacatedRoom := *currentRoom
			vacatedRoom.Status = model.RoomStatusVacant
			if err := roomRepo.UpdateRoom(&vacatedRoom); err != nil {
				return err
			}
			occupiedRoom := *selectedRoom
			occupiedRoom.Status = model.RoomStatusOccupied
			if err := roomRepo.UpdateRoom(&occupiedRoom); err != nil {
				return err
			}
		}
		tenant, err = tenantRepo.GetTenant(id)
		return err
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
			return fmt.Errorf("租客不是在租状态")
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

func tenantWithEdits(current model.Tenant, edited model.Tenant) model.Tenant {
	return model.Tenant{
		ID:               current.ID,
		Name:             edited.Name,
		Phone:            edited.Phone,
		EmergencyContact: edited.EmergencyContact,
		Gender:           edited.Gender,
		RoomID:           edited.RoomID,
		CheckinDate:      edited.CheckinDate,
		LeaseEndDate:     edited.LeaseEndDate,
		CheckoutDate:     current.CheckoutDate,
		RentPrice:        edited.RentPrice,
		RentType:         edited.RentType,
		PaymentTerms:     edited.PaymentTerms,
		Deposit:          edited.Deposit,
		Notes:            edited.Notes,
		Status:           current.Status,
		CreatedAt:        current.CreatedAt,
		UpdatedAt:        current.UpdatedAt,
	}
}

func tenantInputWithRoomDefaults(input TenantInput, room model.Room) TenantInput {
	resolvedInput := input
	if strings.TrimSpace(resolvedInput.RentType) == "" {
		resolvedInput.RentType = model.RentTypeOrDefault(room.RentType)
	}
	if strings.TrimSpace(resolvedInput.PaymentTerms) == "" {
		resolvedInput.PaymentTerms = model.PaymentTermsOrDefault(room.PaymentTerms)
	}
	if strings.TrimSpace(resolvedInput.RentPriceYuan) == "" {
		resolvedInput.RentPriceYuan = fmt.Sprintf("%d", model.RoomRentPrice(room)/100)
	}
	if strings.TrimSpace(resolvedInput.DepositYuan) == "" {
		resolvedInput.DepositYuan = fmt.Sprintf("%d", room.Deposit/100)
	}
	return resolvedInput
}

func buildTenant(input TenantInput) (*model.Tenant, error) {
	name, err := validateName(input.Name)
	if err != nil {
		return nil, err
	}
	phone, err := validatePhone(input.Phone, true, "手机号")
	if err != nil {
		return nil, err
	}
	emergencyContact, err := validatePhone(input.EmergencyContact, false, "紧急联系人")
	if err != nil {
		return nil, err
	}
	gender := strings.TrimSpace(input.Gender)
	if !model.ValidTenantGender(gender) {
		return nil, fmt.Errorf("性别不正确")
	}
	if input.RoomID == 0 {
		return nil, fmt.Errorf("请选择入住房源")
	}
	rentPrice, err := ParseIntegerYuanToFen(input.RentPriceYuan)
	if err != nil {
		return nil, fmt.Errorf("租金金额不正确：%w", err)
	}
	if rentPrice <= 0 {
		return nil, fmt.Errorf("租金金额需大于 0")
	}
	deposit, err := ParseIntegerYuanToFen(input.DepositYuan)
	if err != nil {
		return nil, fmt.Errorf("押金金额不正确：%w", err)
	}
	if deposit < 0 {
		return nil, fmt.Errorf("押金需大于或等于 0")
	}
	rentType := strings.TrimSpace(input.RentType)
	if rentType == "" {
		rentType = model.RentTypeMonthly
	}
	if !model.ValidRentType(rentType) {
		return nil, fmt.Errorf("租金类型不正确")
	}
	paymentTerms := strings.TrimSpace(input.PaymentTerms)
	if paymentTerms == "" {
		paymentTerms = model.PaymentTerms1M1D
	}
	if !model.ValidPaymentTerms(paymentTerms) {
		return nil, fmt.Errorf("付款方式不正确")
	}
	notes, err := validateNotes(input.Notes, "备注")
	if err != nil {
		return nil, err
	}
	checkinDate := input.CheckinDate
	if checkinDate.IsZero() {
		checkinDate = time.Now()
	}
	var leaseEndDate *time.Time
	if !input.LeaseEndDate.IsZero() {
		leaseEnd := input.LeaseEndDate
		if dateOnly(leaseEnd).Before(dateOnly(checkinDate)) {
			return nil, fmt.Errorf("租约到期日不能早于入住日期")
		}
		leaseEndDate = &leaseEnd
	}
	return &model.Tenant{
		Name:             name,
		Phone:            phone,
		EmergencyContact: emergencyContact,
		Gender:           gender,
		RoomID:           input.RoomID,
		CheckinDate:      checkinDate,
		LeaseEndDate:     leaseEndDate,
		RentPrice:        rentPrice,
		RentType:         rentType,
		PaymentTerms:     paymentTerms,
		Deposit:          deposit,
		Notes:            notes,
		Status:           model.TenantStatusActive,
	}, nil
}
