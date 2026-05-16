package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort       string
	AppEnv        string
	SessionSecret string
	DBPath        string
	AdminUsername string
	AdminPassword string
	UploadDir     string
	LandlordName  string
	LandlordPhone string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		AppPort:       getEnv("APP_PORT", "8080"),
		AppEnv:        getEnv("APP_ENV", "development"),
		SessionSecret: getEnv("SESSION_SECRET", "dev-session-secret-change-me"),
		DBPath:        getEnv("DB_PATH", "./data/rent.db"),
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "change-me"),
		UploadDir:     getEnv("UPLOAD_DIR", "./data/uploads"),
		LandlordName:  getEnv("LANDLORD_NAME", "房东"),
		LandlordPhone: getEnv("LANDLORD_PHONE", "13800000000"),
	}

	if cfg.IsProduction() {
		if err := cfg.validateProduction(); err != nil {
			return Config{}, err
		}
	}

	return cfg, nil
}

func (c Config) Addr() string {
	return ":" + c.AppPort
}

func (c Config) IsProduction() bool {
	return strings.EqualFold(c.AppEnv, "production")
}

func (c Config) validateProduction() error {
	missing := make([]string, 0, 3)
	if strings.TrimSpace(os.Getenv("SESSION_SECRET")) == "" {
		missing = append(missing, "SESSION_SECRET")
	}
	if strings.TrimSpace(os.Getenv("ADMIN_USERNAME")) == "" {
		missing = append(missing, "ADMIN_USERNAME")
	}
	if strings.TrimSpace(os.Getenv("ADMIN_PASSWORD")) == "" {
		missing = append(missing, "ADMIN_PASSWORD")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required production config: %s", strings.Join(missing, ", "))
	}
	if c.SessionSecret == "dev-session-secret-change-me" {
		return errors.New("SESSION_SECRET must be changed in production")
	}
	return nil
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
