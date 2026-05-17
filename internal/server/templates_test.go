package server

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/troublesis/rent/internal/model"
)

func TestTemplateDateHelpers(t *testing.T) {
	value := time.Date(2026, time.May, 17, 9, 30, 0, 0, time.Local)
	if got := formatDate(value); got != "2026/05/17" {
		t.Fatalf("formatDate = %q, want 2026/05/17", got)
	}
	if got := formatDateTime(value); got != "2026/05/17 09:30" {
		t.Fatalf("formatDateTime = %q, want 2026/05/17 09:30", got)
	}
	if got := formatInputDate(value); got != "2026-05-17" {
		t.Fatalf("formatInputDate = %q, want 2026-05-17", got)
	}
}

func TestTemplatesParse(t *testing.T) {
	withProjectRoot(t)
	renderer := NewTemplateRenderer("templates")
	tests := []struct {
		layout string
		page   string
	}{
		{layout: "auth_base.html", page: "auth/login.html"},
		{layout: "admin_base.html", page: "admin/dashboard.html"},
		{layout: "admin_base.html", page: "admin/rooms.html"},
		{layout: "admin_base.html", page: "admin/room_form.html"},
		{layout: "admin_base.html", page: "admin/room_detail.html"},
		{layout: "admin_base.html", page: "admin/tenants.html"},
		{layout: "admin_base.html", page: "admin/tenant_form.html"},
		{layout: "admin_base.html", page: "admin/tenant_detail.html"},
		{layout: "admin_base.html", page: "admin/payments.html"},
		{layout: "admin_base.html", page: "admin/stats.html"},
		{layout: "admin_base.html", page: "admin/settings.html"},
		{layout: "public_base.html", page: "public/index.html"},
		{layout: "public_base.html", page: "public/rooms.html"},
		{layout: "public_base.html", page: "public/room_detail.html"},
	}

	for _, tt := range tests {
		t.Run(tt.page, func(t *testing.T) {
			files, err := renderer.templateFiles(tt.layout, tt.page)
			if err != nil {
				t.Fatalf("template files: %v", err)
			}
			if _, err := template.New(tt.layout).Funcs(renderer.funcMap).ParseFiles(files...); err != nil {
				t.Fatalf("parse template: %v", err)
			}
		})
	}
}

func TestRenderPartialExecutesNamedTemplate(t *testing.T) {
	withProjectRoot(t)
	gin.SetMode(gin.TestMode)
	renderer := NewTemplateRenderer("templates")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?partial=1", nil)

	renderer.RenderPartial(c, http.StatusOK, "public/index.html", "public_room_page", gin.H{
		"ViewMode":        "list",
		"HasMore":         true,
		"NextPageURL":     "/?page=2&partial=1&view=list",
		"NextPageFullURL": "/?page=2&view=list",
		"Rooms": []model.Room{{
			ID:           1,
			RoomNo:       "A101",
			Title:        "南向单间",
			Description:  "采光好，近地铁。",
			RentType:     model.RentTypeMonthly,
			RentPrice:    180000,
			Area:         35,
			Floor:        2,
			Bedrooms:     1,
			LivingRooms:  1,
			Bathrooms:    1,
			Orientation:  "南",
			PaymentTerms: model.PaymentTerms1M1D,
		}},
	})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "A101") {
		t.Fatalf("partial body does not contain room content: %s", w.Body.String())
	}
}
