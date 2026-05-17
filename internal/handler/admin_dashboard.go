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

type dashboardSummaryAPIResponse struct {
	TotalRooms             int64 `json:"total_rooms"`
	VacantRooms            int64 `json:"vacant_rooms"`
	OccupiedRooms          int64 `json:"occupied_rooms"`
	ActiveTenants          int64 `json:"active_tenants"`
	OverdueCheckoutTenants int64 `json:"overdue_checkout_tenants"`
	CurrentMonthCollected  int   `json:"current_month_collected"`
	OutstandingAmount      int   `json:"outstanding_amount"`
	CurrentMonthProjected  int   `json:"current_month_projected"`
	SixMonthProjected      int   `json:"six_month_projected"`
	TwelveMonthProjected   int   `json:"twelve_month_projected"`
}

type projectionAPIResponse struct {
	Period       string               `json:"period"`
	TotalFen     int                  `json:"total_fen"`
	TotalText    string               `json:"total_text"`
	CollectedFen int                  `json:"collected_fen"`
	UnpaidFen    int                  `json:"unpaid_fen"`
	NotDueFen    int                  `json:"not_due_fen"`
	Items        []projectionAPIItem  `json:"items"`
	Months       []projectionAPIMonth `json:"months"`
}

type projectionAPIItem struct {
	TenantID          uint   `json:"tenant_id"`
	TenantName        string `json:"tenant_name"`
	RoomNo            string `json:"room_no"`
	DueDate           string `json:"due_date"`
	AmountFen         int    `json:"amount_fen"`
	AmountText        string `json:"amount_text"`
	Status            string `json:"status"`
	RentTypeLabel     string `json:"rent_type_label"`
	PaymentTermsLabel string `json:"payment_terms_label"`
}

type projectionAPIMonth struct {
	Month       string `json:"month"`
	TotalFen    int    `json:"total_fen"`
	TotalText   string `json:"total_text"`
	TenantCount int    `json:"tenant_count"`
}

func NewAdminDashboardHandler(renderer Renderer, dashboardService *service.DashboardService) *AdminDashboardHandler {
	return &AdminDashboardHandler{renderer: renderer, dashboardService: dashboardService}
}

func (h *AdminDashboardHandler) Dashboard(c *gin.Context) {
	summary, err := h.dashboardService.Summary(time.Now())
	if err != nil {
		c.String(http.StatusInternalServerError, "读取仪表盘失败")
		return
	}
	h.renderer.Render(c, http.StatusOK, "admin_base.html", "admin/dashboard.html", gin.H{
		"Title":   "仪表盘",
		"Summary": summary,
		"Error":   queryError(c),
	})
}

func (h *AdminDashboardHandler) APISummary(c *gin.Context) {
	summary, err := h.dashboardService.Summary(time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取仪表盘失败"})
		return
	}
	c.JSON(http.StatusOK, dashboardSummaryAPIResponse{
		TotalRooms:             summary.TotalRooms,
		VacantRooms:            summary.VacantRooms,
		OccupiedRooms:          summary.OccupiedRooms,
		ActiveTenants:          summary.ActiveTenants,
		OverdueCheckoutTenants: summary.OverdueCheckoutTenants,
		CurrentMonthCollected:  summary.CurrentMonthIncome,
		OutstandingAmount:      summary.UnpaidAmount,
		CurrentMonthProjected:  summary.CurrentMonthReceivable,
		SixMonthProjected:      summary.NextSixMonthsReceivable,
		TwelveMonthProjected:   summary.NextTwelveMonthsReceivable,
	})
}

func (h *AdminDashboardHandler) APIProjection(c *gin.Context) {
	period := c.DefaultQuery("period", "month")
	projection, err := h.dashboardService.Projection(period, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取应收预测失败"})
		return
	}
	items := make([]projectionAPIItem, len(projection.Items))
	for i, item := range projection.Items {
		items[i] = projectionItemToAPI(item)
	}
	months := make([]projectionAPIMonth, len(projection.Months))
	for i, month := range projection.Months {
		months[i] = projectionMonthToAPI(month)
	}
	c.JSON(http.StatusOK, projectionAPIResponse{
		Period:       projection.Period,
		TotalFen:     projection.Total,
		TotalText:    service.FormatFen(projection.Total),
		CollectedFen: projection.Collected,
		UnpaidFen:    projection.Unpaid,
		NotDueFen:    projection.NotDue,
		Items:        items,
		Months:       months,
	})
}

func projectionItemToAPI(item service.ProjectionItem) projectionAPIItem {
	return projectionAPIItem{
		TenantID:          item.TenantID,
		TenantName:        item.TenantName,
		RoomNo:            item.RoomNo,
		DueDate:           formatAPIDate(item.DueDate),
		AmountFen:         item.Amount,
		AmountText:        service.FormatFen(item.Amount),
		Status:            item.Status,
		RentTypeLabel:     rentTypeLabelText(item.RentType),
		PaymentTermsLabel: paymentTermsLabelText(item.PaymentTerms),
	}
}

func projectionMonthToAPI(month service.ProjectionMonth) projectionAPIMonth {
	return projectionAPIMonth{
		Month:       month.Month,
		TotalFen:    month.Total,
		TotalText:   service.FormatFen(month.Total),
		TenantCount: month.TenantCount,
	}
}
