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
	statsService   *service.StatsService
}

type statsProjectionAPIResponse struct {
	Period       string               `json:"period"`
	TotalFen     int                  `json:"total_fen"`
	TotalText    string               `json:"total_text"`
	CollectedFen int                  `json:"collected_fen"`
	UnpaidFen    int                  `json:"unpaid_fen"`
	NotDueFen    int                  `json:"not_due_fen"`
	Months       []projectionAPIMonth `json:"months"`
	Note         string               `json:"note"`
}

func NewAdminStatsHandler(renderer Renderer, paymentService *service.PaymentService, statsService *service.StatsService) *AdminStatsHandler {
	return &AdminStatsHandler{renderer: renderer, paymentService: paymentService, statsService: statsService}
}

func (h *AdminStatsHandler) Page(c *gin.Context) {
	now := time.Now()
	year := now.Year()
	if value := c.Query("year"); value != "" {
		if parsedYear, err := strconv.Atoi(value); err == nil && parsedYear > 2000 && parsedYear < 2100 {
			year = parsedYear
		}
	}
	defaultYear := now.Year()
	yearOptions := make([]int, 0, 4)
	for i := 0; i < 4; i++ {
		yearOptions = append(yearOptions, defaultYear-i)
	}
	h.renderer.Render(c, http.StatusOK, "admin_base.html", "admin/stats.html", gin.H{
		"Title":       "数据统计",
		"Year":        year,
		"DefaultYear": defaultYear,
		"YearOptions": yearOptions,
		"Error":       queryError(c),
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

func (h *AdminStatsHandler) Overview(c *gin.Context) {
	filter, ok := h.statsFilter(c)
	if !ok {
		return
	}
	overview, err := h.statsService.Overview(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取收入总览失败"})
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (h *AdminStatsHandler) MonthlyIncome(c *gin.Context) {
	filter, ok := h.statsFilter(c)
	if !ok {
		return
	}
	report, err := h.statsService.MonthlyIncome(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取月度收入失败"})
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *AdminStatsHandler) MonthlyOccupancy(c *gin.Context) {
	filter, ok := h.statsFilter(c)
	if !ok {
		return
	}
	report, err := h.statsService.MonthlyOccupancy(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取出租率失败"})
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *AdminStatsHandler) Projection(c *gin.Context) {
	period := c.DefaultQuery("period", "12months")
	report, err := h.statsService.Projection(period, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取收益预测失败"})
		return
	}
	months := make([]projectionAPIMonth, len(report.Months))
	for i, month := range report.Months {
		months[i] = projectionMonthToAPI(month)
	}
	c.JSON(http.StatusOK, statsProjectionAPIResponse{
		Period:       report.Period,
		TotalFen:     report.TotalFen,
		TotalText:    service.FormatFen(report.TotalFen),
		CollectedFen: report.CollectedFen,
		UnpaidFen:    report.UnpaidFen,
		NotDueFen:    report.NotDueFen,
		Months:       months,
		Note:         report.Note,
	})
}

func (h *AdminStatsHandler) statsFilter(c *gin.Context) (service.StatsFilter, bool) {
	filter, err := statsFilterFromQuery(c, time.Now())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "筛选条件无效"})
		return service.StatsFilter{}, false
	}
	return filter, true
}

func statsFilterFromQuery(c *gin.Context, now time.Time) (service.StatsFilter, error) {
	startValue := c.Query("start_date")
	endValue := c.Query("end_date")
	if startValue != "" || endValue != "" {
		start, err := time.ParseInLocation("2006-01-02", startValue, now.Location())
		if err != nil {
			return service.StatsFilter{}, err
		}
		end, err := time.ParseInLocation("2006-01-02", endValue, now.Location())
		if err != nil {
			return service.StatsFilter{}, err
		}
		return service.NewDateRangeStatsFilter(start, end, now)
	}
	year := now.Year()
	if value := c.Query("year"); value != "" {
		parsedYear, err := strconv.Atoi(value)
		if err != nil {
			return service.StatsFilter{}, err
		}
		year = parsedYear
	}
	return service.NewYearStatsFilter(year, now)
}
