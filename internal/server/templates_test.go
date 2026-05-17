package server

import (
	"html/template"
	"path/filepath"
	"testing"
	"time"
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
			files := []string{filepath.Join("templates", "layout", tt.layout), filepath.Join("templates", tt.page)}
			if _, err := template.New(tt.layout).Funcs(renderer.funcMap).ParseFiles(files...); err != nil {
				t.Fatalf("parse template: %v", err)
			}
		})
	}
}
