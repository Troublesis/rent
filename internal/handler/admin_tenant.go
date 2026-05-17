package handler

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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

type tenantSortLink struct {
	Key       string
	Label     string
	URL       string
	Indicator string
}

type tenantStatusLink struct {
	Value  string
	Label  string
	URL    string
	Active bool
}

type tenantAPIItem struct {
	ID                uint   `json:"id"`
	Name              string `json:"name"`
	Phone             string `json:"phone"`
	RoomID            uint   `json:"room_id"`
	RoomNo            string `json:"room_no"`
	CheckinDate       string `json:"checkin_date"`
	LeaseEndDate      string `json:"lease_end_date"`
	OverdueDays       int    `json:"overdue_days"`
	RentPriceFen      int    `json:"rent_price_fen"`
	RentPriceText     string `json:"rent_price_text"`
	RentType          string `json:"rent_type"`
	RentTypeLabel     string `json:"rent_type_label"`
	PaymentTerms      string `json:"payment_terms"`
	PaymentTermsLabel string `json:"payment_terms_label"`
	DetailURL         string `json:"detail_url"`
	CheckoutURL       string `json:"checkout_url"`
}

func NewAdminTenantHandler(renderer Renderer, tenantService *service.TenantService, roomService *service.RoomService) *AdminTenantHandler {
	return &AdminTenantHandler{renderer: renderer, tenantService: tenantService, roomService: roomService}
}

const (
	adminTenantViewList = "list"
	adminTenantViewCard = "card"
)

func adminTenantViewFromQuery(c *gin.Context) string {
	if c.Query("view") == adminTenantViewCard {
		return adminTenantViewCard
	}
	return adminTenantViewList
}

func tenantStatusQueryValue(filter repository.TenantFilter) string {
	switch filter.Status {
	case "":
		return "all"
	case model.TenantStatusCheckout:
		return model.TenantStatusCheckout
	default:
		return model.TenantStatusActive
	}
}

func tenantClearURL(viewMode string) string {
	if viewMode == "" {
		return "/admin/tenants"
	}
	return "/admin/tenants?view=" + url.QueryEscape(viewMode)
}

func (h *AdminTenantHandler) List(c *gin.Context) {
	filter := tenantFilterFromQuery(c)
	viewMode := adminTenantViewFromQuery(c)
	tenants, err := h.tenantService.ListTenants(filter)
	if err != nil {
		c.String(http.StatusInternalServerError, "读取租客失败")
		return
	}
	h.renderer.Render(c, http.StatusOK, "admin_base.html", "admin/tenants.html", gin.H{
		"Title":          "租客管理",
		"Tenants":        tenants,
		"Statuses":       tenantStatusOptions(),
		"StatusLinks":    tenantStatusLinks(c, filter),
		"SortLinks":      tenantSortLinks(c, filter),
		"Filter":         filter,
		"CurrentStatus":  tenantStatusQueryValue(filter),
		"ViewMode":       viewMode,
		"ListViewURL":    tenantListURL(c, map[string]string{"view": adminTenantViewList}),
		"CardViewURL":    tenantListURL(c, map[string]string{"view": adminTenantViewCard}),
		"ClearFilterURL": tenantClearURL(viewMode),
		"Error":          queryError(c),
	})
}

func (h *AdminTenantHandler) APIList(c *gin.Context) {
	filter := repository.TenantFilter{Status: tenantStatusFromQuery(c.Query("status")), Query: c.Query("q")}
	leaseExpired := c.Query("lease_expired") == "true"
	if leaseExpired {
		filter.Status = model.TenantStatusActive
	}
	tenants, err := h.tenantService.ListTenants(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取租客失败"})
		return
	}
	items := make([]tenantAPIItem, 0, len(tenants))
	now := time.Now()
	for _, tenant := range tenants {
		if leaseExpired && !tenantLeaseExpired(tenant, now) {
			continue
		}
		items = append(items, tenantToAPIItem(tenant, now))
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *AdminTenantHandler) APIActive(c *gin.Context) {
	tenants, err := h.tenantService.ListTenants(repository.TenantFilter{Status: model.TenantStatusActive})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取租客失败"})
		return
	}
	items := make([]tenantAPIItem, len(tenants))
	now := time.Now()
	for i, tenant := range tenants {
		items[i] = tenantToAPIItem(tenant, now)
	}
	c.JSON(http.StatusOK, items)
}

func (h *AdminTenantHandler) New(c *gin.Context) {
	h.renderForm(c, http.StatusOK, service.TenantInput{RentType: model.RentTypeMonthly, PaymentTerms: model.PaymentTerms1M1D}, "/admin/tenants/checkin", "办理入住", "确认入住", "/admin/tenants", "")
}

func (h *AdminTenantHandler) CheckIn(c *gin.Context) {
	input, err := tenantInputFromForm(c)
	if err != nil {
		h.renderForm(c, http.StatusBadRequest, input, "/admin/tenants/checkin", "办理入住", "确认入住", "/admin/tenants", "表单数据不正确")
		return
	}
	if _, err := h.tenantService.CheckInTenant(input); err != nil {
		h.renderForm(c, http.StatusBadRequest, input, "/admin/tenants/checkin", "办理入住", "确认入住", "/admin/tenants", userFacingError(err))
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

func (h *AdminTenantHandler) Edit(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	tenant, err := h.tenantService.GetTenant(id)
	if err != nil {
		c.String(http.StatusNotFound, "租客不存在")
		return
	}
	if tenant.Status != model.TenantStatusActive {
		redirectWithError(c, "/admin/tenants/"+strconv.FormatUint(uint64(id), 10), "只能编辑在租租客")
		return
	}
	detailURL := "/admin/tenants/" + strconv.FormatUint(uint64(id), 10)
	h.renderForm(c, http.StatusOK, tenantInputFromTenant(*tenant), detailURL, "编辑租客", "保存修改", detailURL, "")
}

func (h *AdminTenantHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	detailURL := "/admin/tenants/" + strconv.FormatUint(uint64(id), 10)
	input, err := tenantInputFromForm(c)
	if err != nil {
		h.renderForm(c, http.StatusBadRequest, input, detailURL, "编辑租客", "保存修改", detailURL, "表单数据不正确")
		return
	}
	if _, err := h.tenantService.UpdateTenant(id, input); err != nil {
		h.renderForm(c, http.StatusBadRequest, input, detailURL, "编辑租客", "保存修改", detailURL, userFacingError(err))
		return
	}
	c.Redirect(http.StatusSeeOther, detailURL)
}

func (h *AdminTenantHandler) CheckOut(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.tenantService.CheckOutTenant(id); err != nil {
		redirectWithError(c, "/admin/tenants/"+strconv.FormatUint(uint64(id), 10), userFacingError(err))
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/tenants/"+strconv.FormatUint(uint64(id), 10))
}

func (h *AdminTenantHandler) renderForm(c *gin.Context, status int, input service.TenantInput, action string, title string, submitLabel string, backURL string, errorMessage string) {
	rooms, err := h.tenantFormRooms(input.RoomID)
	if err != nil {
		c.String(http.StatusInternalServerError, "读取房源失败")
		return
	}
	if input.RentType == "" {
		input.RentType = model.RentTypeMonthly
	}
	if input.PaymentTerms == "" {
		input.PaymentTerms = model.PaymentTerms1M1D
	}
	h.renderer.Render(c, status, "admin_base.html", "admin/tenant_form.html", gin.H{
		"Title":        title,
		"Input":        input,
		"Rooms":        rooms,
		"Action":       action,
		"SubmitLabel":  submitLabel,
		"BackURL":      backURL,
		"RentTypes":    rentTypeOptions(),
		"PaymentTerms": paymentTermsOptions(),
		"Genders":      tenantGenderOptions(),
		"Error":        errorMessage,
	})
}

func (h *AdminTenantHandler) tenantFormRooms(selectedRoomID uint) ([]model.Room, error) {
	rooms, err := h.roomService.ListRooms(repository.RoomFilter{Status: model.RoomStatusVacant})
	if err != nil {
		return nil, err
	}
	if selectedRoomID == 0 {
		return rooms, nil
	}
	for _, room := range rooms {
		if room.ID == selectedRoomID {
			return rooms, nil
		}
	}
	selectedRoom, err := h.roomService.GetRoom(selectedRoomID)
	if err != nil {
		return rooms, nil
	}
	return append([]model.Room{*selectedRoom}, rooms...), nil
}

func tenantInputFromTenant(tenant model.Tenant) service.TenantInput {
	leaseEndDate := time.Time{}
	if tenant.LeaseEndDate != nil {
		leaseEndDate = *tenant.LeaseEndDate
	}
	return service.TenantInput{
		Name:             tenant.Name,
		Phone:            tenant.Phone,
		EmergencyContact: tenant.EmergencyContact,
		Gender:           tenant.Gender,
		RoomID:           tenant.RoomID,
		CheckinDate:      tenant.CheckinDate,
		LeaseEndDate:     leaseEndDate,
		RentPriceYuan:    service.FormatFenAsYuanInt(tenant.RentPrice),
		RentType:         tenant.RentType,
		PaymentTerms:     tenant.PaymentTerms,
		DepositYuan:      service.FormatFenAsYuanInt(tenant.Deposit),
		Notes:            tenant.Notes,
	}
}

func tenantInputFromForm(c *gin.Context) (service.TenantInput, error) {
	input := service.TenantInput{
		Name:             c.PostForm("name"),
		Phone:            c.PostForm("phone"),
		EmergencyContact: c.PostForm("emergency_contact"),
		Gender:           c.PostForm("gender"),
		RentPriceYuan:    c.PostForm("rent_price"),
		RentType:         c.PostForm("rent_type"),
		PaymentTerms:     c.PostForm("payment_terms"),
		DepositYuan:      c.PostForm("deposit"),
		Notes:            c.PostForm("notes"),
	}
	roomID, err := parseUintForm(c, "room_id")
	if err != nil {
		return input, err
	}
	input.RoomID = roomID
	checkinDate, err := parseDateForm(c, "checkin_date")
	if err != nil {
		return input, err
	}
	input.CheckinDate = checkinDate
	leaseEndDate, err := parseDateForm(c, "lease_end_date")
	if err != nil {
		return input, err
	}
	input.LeaseEndDate = leaseEndDate
	return input, nil
}

func tenantFilterFromQuery(c *gin.Context) repository.TenantFilter {
	return repository.TenantFilter{
		Status:  tenantStatusFromQuery(c.Query("status")),
		Query:   strings.TrimSpace(c.Query("q")),
		SortBy:  c.DefaultQuery("sort_by", "created_at"),
		SortDir: c.DefaultQuery("sort_dir", "desc"),
	}
}

func tenantStatusFromQuery(status string) string {
	switch status {
	case "all":
		return ""
	case model.TenantStatusCheckout:
		return model.TenantStatusCheckout
	default:
		return model.TenantStatusActive
	}
}

func tenantSortLinks(c *gin.Context, filter repository.TenantFilter) []tenantSortLink {
	items := []tenantSortLink{
		{Key: "name", Label: "租客"},
		{Key: "room", Label: "房源"},
		{Key: "rent_price", Label: "租金"},
		{Key: "checkin_date", Label: "入住日期"},
		{Key: "status", Label: "状态"},
	}
	for i := range items {
		items[i].URL = tenantListURL(c, map[string]string{"sort_by": items[i].Key, "sort_dir": nextTenantSortDir(filter, items[i].Key)})
		if filter.SortBy == items[i].Key {
			items[i].Indicator = "↓"
			if strings.EqualFold(filter.SortDir, "asc") {
				items[i].Indicator = "↑"
			}
		}
	}
	return items
}

func nextTenantSortDir(filter repository.TenantFilter, key string) string {
	if filter.SortBy == key && strings.EqualFold(filter.SortDir, "asc") {
		return "desc"
	}
	return "asc"
}

func tenantStatusLinks(c *gin.Context, filter repository.TenantFilter) []tenantStatusLink {
	items := []tenantStatusLink{
		{Value: "all", Label: "全部", Active: filter.Status == ""},
		{Value: model.TenantStatusActive, Label: "在租", Active: filter.Status == model.TenantStatusActive},
		{Value: model.TenantStatusCheckout, Label: "已退租", Active: filter.Status == model.TenantStatusCheckout},
	}
	for i := range items {
		items[i].URL = tenantListURL(c, map[string]string{"status": items[i].Value})
	}
	return items
}

func tenantListURL(c *gin.Context, overrides map[string]string) string {
	values := url.Values{}
	for key, vals := range c.Request.URL.Query() {
		for _, value := range vals {
			values.Add(key, value)
		}
	}
	for key, value := range overrides {
		values.Set(key, value)
	}
	encoded := values.Encode()
	if encoded == "" {
		return "/admin/tenants"
	}
	return "/admin/tenants?" + encoded
}

func tenantToAPIItem(tenant model.Tenant, now time.Time) tenantAPIItem {
	overdueDays := 0
	if tenantLeaseExpired(tenant, now) {
		overdueDays = int(paymentAPIDateOnly(now).Sub(paymentAPIDateOnly(*tenant.LeaseEndDate)).Hours() / 24)
	}
	return tenantAPIItem{
		ID:                tenant.ID,
		Name:              tenant.Name,
		Phone:             tenant.Phone,
		RoomID:            tenant.RoomID,
		RoomNo:            tenant.Room.RoomNo,
		CheckinDate:       formatAPIDate(tenant.CheckinDate),
		LeaseEndDate:      formatOptionalAPIDate(tenant.LeaseEndDate),
		OverdueDays:       overdueDays,
		RentPriceFen:      tenant.RentPrice,
		RentPriceText:     service.FormatFenAsYuanInt(tenant.RentPrice),
		RentType:          tenant.RentType,
		RentTypeLabel:     rentTypeLabelText(tenant.RentType),
		PaymentTerms:      tenant.PaymentTerms,
		PaymentTermsLabel: paymentTermsLabelText(tenant.PaymentTerms),
		DetailURL:         "/admin/tenants/" + strconv.FormatUint(uint64(tenant.ID), 10),
		CheckoutURL:       "/admin/tenants/" + strconv.FormatUint(uint64(tenant.ID), 10),
	}
}

func tenantLeaseExpired(tenant model.Tenant, now time.Time) bool {
	if tenant.Status != model.TenantStatusActive || tenant.LeaseEndDate == nil {
		return false
	}
	return paymentAPIDateOnly(*tenant.LeaseEndDate).Before(paymentAPIDateOnly(now))
}

func tenantStatusOptions() []SelectOption {
	return []SelectOption{
		{Value: model.TenantStatusActive, Label: "在租"},
		{Value: model.TenantStatusCheckout, Label: "已退租"},
	}
}

func tenantGenderOptions() []SelectOption {
	return []SelectOption{
		{Value: model.TenantGenderMale, Label: "男性"},
		{Value: model.TenantGenderFemale, Label: "女性"},
	}
}
