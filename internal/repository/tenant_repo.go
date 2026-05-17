package repository

import (
	"github.com/troublesis/rent/internal/model"
	"gorm.io/gorm"
)

type TenantFilter struct {
	Status string
	Query  string
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
	query := r.db.Model(&model.Tenant{}).Preload("Room").Order("created_at DESC")
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Query != "" {
		like := "%" + filter.Query + "%"
		query = query.Where("name LIKE ? OR phone LIKE ?", like, like)
	}
	var tenants []model.Tenant
	if err := query.Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
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
