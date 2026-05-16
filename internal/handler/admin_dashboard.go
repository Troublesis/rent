package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/troublesis/rent/internal/service"
)

type AdminDashboardHandler struct {
	renderer         Renderer
	dashboardService *service.DashboardService
}

func NewAdminDashboardHandler(renderer Renderer, dashboardService *service.DashboardService) *AdminDashboardHandler {
	return &AdminDashboardHandler{renderer: renderer, dashboardService: dashboardService}
}

func (h *AdminDashboardHandler) Dashboard(c *gin.Context) {
	summary, err := h.dashboardService.Summary(time.Now())
	if err != nil {
		c.String(http.StatusInternalServerError, "读取仪表盘失败: %v", err)
		return
	}
	h.renderer.Render(c, http.StatusOK, "admin_base.html", "admin/dashboard.html", gin.H{
		"Title":   "仪表盘",
		"Summary": summary,
		"Error":   queryError(c),
	})
}
