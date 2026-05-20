package repository

import (
	"strings"
	"time"

	"github.com/troublesis/rent/internal/model"
	"gorm.io/gorm"
)

type TenantFilter struct {
	Status  string
	Query   string
	SortBy  string
	SortDir string
}

type TenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) WithDB(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) ListTenants(filter TenantFilter) ([]model.Tenant, error) {
	query := r.db.Model(&model.Tenant{}).Preload("Room")
	joinedRooms := false
	if filter.Status != "" {
		query = query.Where("tenants.status = ?", filter.Status)
	}
	if strings.TrimSpace(filter.Query) != "" {
		like := "%" + strings.TrimSpace(filter.Query) + "%"
		query = query.Joins("LEFT JOIN rooms ON rooms.id = tenants.room_id")
		joinedRooms = true
		query = query.Where("tenants.name LIKE ? OR tenants.phone LIKE ? OR rooms.room_no LIKE ? OR rooms.title LIKE ? OR (tenants.name || ' - ' || rooms.room_no || ' - ' || tenants.phone) LIKE ?", like, like, like, like, like)
	}
	orderColumn, orderDirection, needsRoomJoin := tenantSort(filter.SortBy, filter.SortDir)
	if needsRoomJoin && !joinedRooms {
		query = query.Joins("LEFT JOIN rooms ON rooms.id = tenants.room_id")
	}
	query = query.Order(orderColumn + " " + orderDirection)
	var tenants []model.Tenant
	if err := query.Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

func tenantSort(sortBy string, sortDir string) (string, string, bool) {
	direction := "DESC"
	if strings.EqualFold(sortDir, "asc") {
		direction = "ASC"
	}
	switch sortBy {
	case "name":
		return "tenants.name", direction, false
	case "room":
		return "rooms.room_no", direction, true
	case "rent_price":
		return "tenants.rent_price", direction, false
	case "checkin_date":
		return "tenants.checkin_date", direction, false
	case "status":
		return "tenants.status", direction, false
	default:
		return "tenants.created_at", "DESC", false
	}
}

func (r *TenantRepository) ListActiveTenantsWithPayments() ([]model.Tenant, error) {
	var tenants []model.Tenant
	if err := r.db.Model(&model.Tenant{}).
		Preload("Room").
		Preload("Payments", func(db *gorm.DB) *gorm.DB {
			return db.Order("pay_date ASC, created_at ASC")
		}).
		Where("status = ?", model.TenantStatusActive).
		Order("checkin_date ASC, created_at ASC").
		Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

func (r *TenantRepository) GetTenant(id uint) (*model.Tenant, error) {
	var tenant model.Tenant
	if err := r.db.Preload("Room").Preload("Payments", func(db *gorm.DB) *gorm.DB {
		return db.Order("pay_date DESC, created_at DESC")
	}).First(&tenant, id).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepository) GetActiveTenantByRoomID(roomID uint) (*model.Tenant, error) {
	var tenant model.Tenant
	if err := r.db.Where("room_id = ? AND status = ?", roomID, model.TenantStatusActive).First(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepository) CreateTenant(tenant *model.Tenant) error {
	return r.db.Create(tenant).Error
}

func (r *TenantRepository) UpdateTenant(tenant *model.Tenant) error {
	return r.db.Save(tenant).Error
}

func (r *TenantRepository) CountActiveTenants() (int64, error) {
	var count int64
	if err := r.db.Model(&model.Tenant{}).Where("status = ?", model.TenantStatusActive).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *TenantRepository) ListTenantsOverlappingPeriod(start time.Time, end time.Time) ([]model.Tenant, error) {
	var tenants []model.Tenant
	if err := r.db.Model(&model.Tenant{}).
		Preload("Room").
		Where("checkin_date < ?", end).
		Where("checkout_date IS NULL OR checkout_date >= ?", start).
		Order("checkin_date ASC, created_at ASC").
		Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}
