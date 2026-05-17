package service

import (
	"fmt"
	"strings"

	"github.com/troublesis/rent/config"
	"github.com/troublesis/rent/internal/repository"
)

const (
	SettingLandlordName  = "landlord_name"
	SettingLandlordPhone = "landlord_phone"
)

type Settings struct {
	LandlordName  string
	LandlordPhone string
}

type SettingsService struct {
	cfg          config.Config
	settingsRepo *repository.SettingsRepository
}

func NewSettingsService(cfg config.Config, settingsRepo *repository.SettingsRepository) *SettingsService {
	return &SettingsService{cfg: cfg, settingsRepo: settingsRepo}
}

func (s *SettingsService) GetSettings() (Settings, error) {
	settings, err := s.settingsRepo.All()
	if err != nil {
		return Settings{}, err
	}
	return Settings{
		LandlordName:  valueOrDefault(settings[SettingLandlordName], s.cfg.LandlordName),
		LandlordPhone: valueOrDefault(settings[SettingLandlordPhone], s.cfg.LandlordPhone),
	}, nil
}

func (s *SettingsService) UpdateSettings(settings Settings) error {
	landlordName := strings.TrimSpace(settings.LandlordName)
	if landlordName == "" {
		return fmt.Errorf("房东姓名不能为空")
	}
	landlordPhone, err := validatePhone(settings.LandlordPhone, true, "联系电话")
	if err != nil {
		return err
	}
	if err := s.settingsRepo.Set(SettingLandlordName, landlordName); err != nil {
		return err
	}
	return s.settingsRepo.Set(SettingLandlordPhone, landlordPhone)
}

func valueOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
