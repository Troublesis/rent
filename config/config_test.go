package config

import (
	"strings"
	"testing"
)

func TestLoadUsesDevelopmentDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("APP_PORT", "")
	t.Setenv("SESSION_SECRET", "")
	t.Setenv("ADMIN_USERNAME", "")
	t.Setenv("ADMIN_PASSWORD", "")
	t.Setenv("DB_PATH", "")
	t.Setenv("APP_TIMEZONE", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.AppPort != "8080" {
		t.Fatalf("AppPort = %q, want 8080", cfg.AppPort)
	}
	if cfg.Addr() != ":8080" {
		t.Fatalf("Addr = %q, want :8080", cfg.Addr())
	}
	if cfg.IsProduction() {
		t.Fatal("development config should not be production")
	}
	if cfg.AppTimezone != "Asia/Shanghai" {
		t.Fatalf("AppTimezone = %q, want Asia/Shanghai", cfg.AppTimezone)
	}
}

func TestConfigLocationLoadsTimezone(t *testing.T) {
	cfg := Config{AppTimezone: "UTC"}
	loc, err := cfg.Location()
	if err != nil {
		t.Fatalf("Location returned error: %v", err)
	}
	if loc.String() != "UTC" {
		t.Fatalf("location = %q, want UTC", loc.String())
	}
}

func TestConfigLocationRejectsInvalidTimezone(t *testing.T) {
	cfg := Config{AppTimezone: "Invalid/Timezone"}
	if _, err := cfg.Location(); err == nil {
		t.Fatal("Location returned nil error for invalid timezone")
	}
}

func TestLoadRequiresProductionSecrets(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("SESSION_SECRET", "")
	t.Setenv("ADMIN_USERNAME", "")
	t.Setenv("ADMIN_PASSWORD", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load returned nil error for missing production secrets")
	}
	if !strings.Contains(err.Error(), "SESSION_SECRET") {
		t.Fatalf("error = %q, want SESSION_SECRET mention", err.Error())
	}
}
