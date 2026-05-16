package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
	"github.com/troublesis/rent/internal/service"
)

type AdminPaymentHandler struct {
	renderer       Renderer
	paymentService *service.PaymentService
	tenantService  *service.TenantService
}

func NewAdminPaymentHandler(renderer Renderer, paymentService *service.PaymentService, tenantService *service.TenantService) *AdminPaymentHandler {
	return &AdminPaymentHandler{renderer: renderer, paymentService: paymentService, tenantService: tenantService}
}

func (h *AdminPaymentHandler) List(c *gin.Context) {
	filter := paymentFilterFromQuery(c)
	payments, err := h.paymentService.ListPayments(filter)
	if err != nil {
		c.String(http.StatusInternalServerError, "读取收款记录失败: %v", err)
		return
	}
	tenants, err := h.tenantService.ListTenants(repository.TenantFilter{Status: model.TenantStatusActive})
	if err != nil {
		c.String(http.StatusInternalServerError, "读取租客失败: %v", err)
		return
	}
	h.renderer.Render(c, http.StatusOK, "admin_base.html", "admin/payments.html", gin.H{
		"Title":        "收款记录",
		"Payments":     payments,
		"Tenants":      tenants,
		"Filter":       filter,
		"PaymentTypes": paymentTypeOptions(),
		"Error":        queryError(c),
	})
}

func (h *AdminPaymentHandler) Create(c *gin.Context) {
	input, err := paymentInputFromForm(c)
	if err != nil {
		redirectWithError(c, "/admin/payments", "表单数据不正确")
		return
	}
	if _, err := h.paymentService.RecordPayment(input); err != nil {
		redirectWithError(c, "/admin/payments", err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/payments")
}

func (h *AdminPaymentHandler) Toggle(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.paymentService.TogglePaid(id); err != nil {
		redirectWithError(c, "/admin/payments", err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/payments")
}

func paymentFilterFromQuery(c *gin.Context) repository.PaymentFilter {
	filter := repository.PaymentFilter{Type: c.Query("type")}
	switch c.Query("paid") {
	case "true":
		paid := true
		filter.Paid = &paid
	case "false":
		paid := false
		filter.Paid = &paid
	}
	if tenantID, err := strconv.ParseUint(c.Query("tenant_id"), 10, 64); err == nil {
		filter.TenantID = uint(tenantID)
	}
	return filter
}

func paymentInputFromForm(c *gin.Context) (service.PaymentInput, error) {
	tenantID, err := parseUintForm(c, "tenant_id")
	if err != nil {
		return service.PaymentInput{}, err
	}
	payDate, err := parseDateForm(c, "pay_date")
	if err != nil {
		return service.PaymentInput{}, err
	}
	return service.PaymentInput{
		TenantID:   tenantID,
		AmountYuan: c.PostForm("amount"),
		Type:       c.PostForm("type"),
		Paid:       c.PostForm("paid") == "on",
		PayDate:    payDate,
		Note:       c.PostForm("note"),
	}, nil
}

func paymentTypeOptions() []SelectOption {
	return []SelectOption{
		{Value: model.PaymentTypeRent, Label: "租金"},
		{Value: model.PaymentTypeWater, Label: "水费"},
		{Value: model.PaymentTypeElectricity, Label: "电费"},
		{Value: model.PaymentTypeOther, Label: "其他"},
	}
}
