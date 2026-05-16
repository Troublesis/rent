package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
	"github.com/troublesis/rent/internal/service"
)

type AdminTenantHandler struct {
	renderer      Renderer
	tenantService *service.TenantService
	roomService   *service.RoomService
}

func NewAdminTenantHandler(renderer Renderer, tenantService *service.TenantService, roomService *service.RoomService) *AdminTenantHandler {
	return &AdminTenantHandler{renderer: renderer, tenantService: tenantService, roomService: roomService}
}

func (h *AdminTenantHandler) List(c *gin.Context) {
	filter := repository.TenantFilter{Status: c.Query("status"), Query: c.Query("q")}
	tenants, err := h.tenantService.ListTenants(filter)
	if err != nil {
		c.String(http.StatusInternalServerError, "读取租客失败: %v", err)
		return
	}
	h.renderer.Render(c, http.StatusOK, "admin_base.html", "admin/tenants.html", gin.H{
		"Title":    "租客管理",
		"Tenants":  tenants,
		"Statuses": tenantStatusOptions(),
		"Filter":   filter,
		"Error":    queryError(c),
	})
}

func (h *AdminTenantHandler) New(c *gin.Context) {
	h.renderForm(c, http.StatusOK, service.TenantInput{}, "")
}

func (h *AdminTenantHandler) CheckIn(c *gin.Context) {
	input, err := tenantInputFromForm(c)
	if err != nil {
		h.renderForm(c, http.StatusBadRequest, input, "表单数据不正确")
		return
	}
	if _, err := h.tenantService.CheckInTenant(input); err != nil {
		h.renderForm(c, http.StatusBadRequest, input, err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/tenants")
}

func (h *AdminTenantHandler) Detail(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	tenant, err := h.tenantService.GetTenant(id)
	if err != nil {
		c.String(http.StatusNotFound, "租客不存在")
		return
	}
	h.renderer.Render(c, http.StatusOK, "admin_base.html", "admin/tenant_detail.html", gin.H{
		"Title":  "租客详情",
		"Tenant": tenant,
		"Error":  queryError(c),
	})
}

func (h *AdminTenantHandler) CheckOut(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.tenantService.CheckOutTenant(id); err != nil {
		redirectWithError(c, "/admin/tenants/"+strconv.FormatUint(uint64(id), 10), err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/tenants/"+strconv.FormatUint(uint64(id), 10))
}

func (h *AdminTenantHandler) renderForm(c *gin.Context, status int, input service.TenantInput, errorMessage string) {
	rooms, err := h.roomService.ListRooms(repository.RoomFilter{Status: model.RoomStatusVacant})
	if err != nil {
		c.String(http.StatusInternalServerError, "读取空置房源失败: %v", err)
		return
	}
	h.renderer.Render(c, status, "admin_base.html", "admin/tenant_form.html", gin.H{
		"Title": "办理入住",
		"Input": input,
		"Rooms": rooms,
		"Error": errorMessage,
	})
}

func tenantInputFromForm(c *gin.Context) (service.TenantInput, error) {
	roomID, err := parseUintForm(c, "room_id")
	if err != nil {
		return service.TenantInput{}, err
	}
	checkinDate, err := parseDateForm(c, "checkin_date")
	if err != nil {
		return service.TenantInput{}, err
	}
	return service.TenantInput{
		Name:             c.PostForm("name"),
		Phone:            c.PostForm("phone"),
		EmergencyContact: c.PostForm("emergency_contact"),
		RoomID:           roomID,
		CheckinDate:      checkinDate,
		RentPriceYuan:    c.PostForm("rent_price"),
		DepositYuan:      c.PostForm("deposit"),
	}, nil
}

func tenantStatusOptions() []SelectOption {
	return []SelectOption{
		{Value: model.TenantStatusActive, Label: "在租"},
		{Value: model.TenantStatusCheckout, Label: "已退租"},
	}
}
