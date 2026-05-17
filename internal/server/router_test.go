package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/troublesis/rent/config"
	"github.com/troublesis/rent/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestProtectedAdminRouteRedirectsToLogin(t *testing.T) {
	withProjectRoot(t)
	router := newTestRouter(t)

	request := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusSeeOther)
	}
	if location := response.Header().Get("Location"); location != "/admin/login" {
		t.Fatalf("Location = %q, want /admin/login", location)
	}
}

func TestProtectedPaymentPatchRouteRedirectsToLogin(t *testing.T) {
	withProjectRoot(t)
	router := newTestRouter(t)

	request := httptest.NewRequest(http.MethodPatch, "/admin/payments/1/exclude", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusSeeOther)
	}
	if location := response.Header().Get("Location"); location != "/admin/login" {
		t.Fatalf("Location = %q, want /admin/login", location)
	}
}

func TestLoginPageReturnsOK(t *testing.T) {
	withProjectRoot(t)
	router := newTestRouter(t)

	request := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	if !strings.Contains(body, "房东登录") {
		t.Fatal("login page should contain Chinese title")
	}
	if !strings.Contains(body, "记住密码") {
		t.Fatal("login page should use updated remember wording")
	}
	if strings.Contains(body, "记住我 30 天") {
		t.Fatal("login page should not mention day-based remember duration")
	}
}

func TestPublicPageDoesNotExposeAdminLogin(t *testing.T) {
	withProjectRoot(t)
	router := newTestRouter(t)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	if strings.Contains(body, "房东登录") || strings.Contains(body, "/admin/login") {
		t.Fatal("public page should not expose admin login")
	}
}

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&model.Room{}, &model.RoomMedia{}, &model.Tenant{}, &model.Payment{}, &model.AppSetting{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	return NewRouter(config.Config{
		AppPort:       "8080",
		AppEnv:        "test",
		SessionSecret: "test-session-secret",
		DBPath:        ":memory:",
		AdminUsername: "admin",
		AdminPassword: "password",
		UploadDir:     "./data/uploads",
		LandlordName:  "房东",
		LandlordPhone: "13800000000",
	}, db)
}

func withProjectRoot(t *testing.T) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("change to project root: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(original)
	})
}
