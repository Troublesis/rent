package server

import (
	"html/template"
	"path/filepath"
	"testing"
)

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
