package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/troublesis/rent/internal/service"
)

type AdminStatsHandler struct {
	renderer       Renderer
	paymentService *service.PaymentService
}

func NewAdminStatsHandler(renderer Renderer, paymentService *service.PaymentService) *AdminStatsHandler {
	return &AdminStatsHandler{renderer: renderer, paymentService: paymentService}
}

func (h *AdminStatsHandler) Page(c *gin.Context) {
	h.renderer.Render(c, http.StatusOK, "admin_base.html", "admin/stats.html", gin.H{
		"Title": "数据统计",
		"Year":  time.Now().Year(),
		"Error": queryError(c),
	})
}

func (h *AdminStatsHandler) DashboardStats(c *gin.Context) {
	year := time.Now().Year()
	if value := c.Query("year"); value != "" {
		parsedYear, err := strconv.Atoi(value)
		if err == nil && parsedYear > 2000 && parsedYear < 2100 {
			year = parsedYear
		}
	}
	rows, err := h.paymentService.MonthlyIncome(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取统计数据失败"})
		return
	}
	totals := make([]int, 12)
	for _, row := range rows {
		if row.Month >= 1 && row.Month <= 12 {
			totals[row.Month-1] = row.Total
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"year":      year,
		"labels":    []string{"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"},
		"totalsFen": totals,
	})
}
