package service

import (
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
	if err := s.settingsRepo.Set(SettingLandlordName, settings.LandlordName); err != nil {
		return err
	}
	return s.settingsRepo.Set(SettingLandlordPhone, settings.LandlordPhone)
}

func valueOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
