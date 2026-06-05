package repository

import (
	"fmt"

	"github.com/troublesis/rent/internal/model"
	"gorm.io/gorm"
)

type PushRepository struct {
	db *gorm.DB
}

func NewPushRepository(db *gorm.DB) *PushRepository {
	return &PushRepository{db: db}
}

func (r *PushRepository) GetConfig() (*model.PushConfig, error) {
	var config model.PushConfig
	if err := r.db.First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *PushRepository) GetConfigOrCreate() (*model.PushConfig, error) {
	config, err := r.GetConfig()
	if err == nil {
		return config, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	defaultConfig := model.PushConfig{ID: 1}
	if err := r.db.Create(&defaultConfig).Error; err != nil {
		return nil, fmt.Errorf("create default push config: %w", err)
	}
	return &defaultConfig, nil
}

func (r *PushRepository) SaveConfig(config *model.PushConfig) error {
	return r.db.Save(config).Error
}

func (r *PushRepository) GetTemplate() (*model.PushTemplate, error) {
	var tmpl model.PushTemplate
	if err := r.db.First(&tmpl).Error; err != nil {
		return nil, err
	}
	return &tmpl, nil
}

func (r *PushRepository) GetTemplateOrCreate() (*model.PushTemplate, error) {
	tmpl, err := r.GetTemplate()
	if err == nil {
		return tmpl, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	defaultTmpl := model.DefaultPushTemplate()
	if err := r.db.Create(&defaultTmpl).Error; err != nil {
		return nil, fmt.Errorf("create default push template: %w", err)
	}
	return &defaultTmpl, nil
}

func (r *PushRepository) SaveTemplate(tmpl *model.PushTemplate) error {
	return r.db.Save(tmpl).Error
}

func (r *PushRepository) CreateLog(log *model.PushLog) error {
	return r.db.Create(log).Error
}
