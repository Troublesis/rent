package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
	"github.com/troublesis/rent/internal/service"
)

type AdminRoomHandler struct {
	renderer      Renderer
	roomService   *service.RoomService
	tenantService *service.TenantService
}

type roomAPIItem struct {
	ID            uint   `json:"id"`
	RoomNo        string `json:"room_no"`
	Title         string `json:"title"`
	Status        string `json:"status"`
	StatusLabel   string `json:"status_label"`
	RentPriceFen  int    `json:"rent_price_fen"`
	RentPriceText string `json:"rent_price_text"`
	RentType      string `json:"rent_type"`
	RentTypeLabel string `json:"rent_type_label"`
	Floor         int    `json:"floor"`
	Area          int    `json:"area"`
	TenantID      uint   `json:"tenant_id"`
	TenantName    string `json:"tenant_name"`
	LeaseEndDate  string `json:"lease_end_date"`
	DetailURL     string `json:"detail_url"`
}

func NewAdminRoomHandler(renderer Renderer, roomService *service.RoomService, tenantService *service.TenantService) *AdminRoomHandler {
	return &AdminRoomHandler{renderer: renderer, roomService: roomService, tenantService: tenantService}
}

func (h *AdminRoomHandler) List(c *gin.Context) {
	filter := repository.RoomFilter{Status: c.Query("status"), Query: c.Query("q")}
	rooms, err := h.roomService.ListRooms(filter)
	if err != nil {
		c.String(http.StatusInternalServerError, "读取房源失败")
		return
	}
	h.renderer.Render(c, http.StatusOK, "admin_base.html", "admin/rooms.html", gin.H{
		"Title":    "房源管理",
		"Rooms":    rooms,
		"Statuses": roomStatusOptions(),
		"Filter":   filter,
		"Error":    queryError(c),
	})
}

func (h *AdminRoomHandler) APIList(c *gin.Context) {
	filter := repository.RoomFilter{Status: c.Query("status"), Query: c.Query("q")}
	if c.Query("include_all") == "true" {
		filter.Status = ""
	}
	rooms, err := h.roomService.ListRooms(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取房源失败"})
		return
	}
	tenantByRoomID := map[uint]model.Tenant{}
	if h.tenantService != nil {
		tenants, err := h.tenantService.ListTenants(repository.TenantFilter{Status: model.TenantStatusActive})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取租客失败"})
			return
		}
		for _, tenant := range tenants {
			tenantByRoomID[tenant.RoomID] = tenant
		}
	}
	items := make([]roomAPIItem, len(rooms))
	for i, room := range rooms {
		items[i] = roomToAPIItem(room, tenantByRoomID[room.ID])
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *AdminRoomHandler) New(c *gin.Context) {
	room := model.Room{Status: model.RoomStatusVacant, RentType: model.RentTypeMonthly, PaymentTerms: model.PaymentTerms1M1D, Area: 1}
	h.renderForm(c, http.StatusOK, room, "/admin/rooms", "新增房源", "")
}

func (h *AdminRoomHandler) Create(c *gin.Context) {
	input, err := roomInputFromForm(c)
	if err != nil {
		h.renderForm(c, http.StatusBadRequest, roomFromInput(input), "/admin/rooms", "新增房源", "表单数据不正确")
		return
	}
	if _, err := h.roomService.CreateRoom(input); err != nil {
		h.renderForm(c, http.StatusBadRequest, roomFromInput(input), "/admin/rooms", "新增房源", userFacingError(err))
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/rooms")
}

func (h *AdminRoomHandler) Detail(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	room, err := h.roomService.GetRoom(id)
	if err != nil {
		c.String(http.StatusNotFound, "房源不存在")
		return
	}
	h.renderer.Render(c, http.StatusOK, "admin_base.html", "admin/room_detail.html", gin.H{
		"Title": "房源详情",
		"Room":  room,
		"Error": queryError(c),
	})
}

func (h *AdminRoomHandler) Edit(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	room, err := h.roomService.GetRoom(id)
	if err != nil {
		c.String(http.StatusNotFound, "房源不存在")
		return
	}
	h.renderForm(c, http.StatusOK, *room, "/admin/rooms/"+strconv.FormatUint(uint64(id), 10), "编辑房源", "")
}

func (h *AdminRoomHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	input, err := roomInputFromForm(c)
	if err != nil {
		h.renderForm(c, http.StatusBadRequest, roomFromInput(input), "/admin/rooms/"+strconv.FormatUint(uint64(id), 10), "编辑房源", "表单数据不正确")
		return
	}
	if _, err := h.roomService.UpdateRoom(id, input); err != nil {
		h.renderForm(c, http.StatusBadRequest, roomFromInput(input), "/admin/rooms/"+strconv.FormatUint(uint64(id), 10), "编辑房源", userFacingError(err))
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/rooms/"+strconv.FormatUint(uint64(id), 10))
}

func (h *AdminRoomHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.roomService.DeleteRoom(id); err != nil {
		redirectWithError(c, "/admin/rooms", userFacingError(err))
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/rooms")
}

func (h *AdminRoomHandler) AddVideoLink(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.roomService.AddRoomVideoLink(id, c.PostForm("video_link")); err != nil {
		redirectWithError(c, "/admin/rooms/"+strconv.FormatUint(uint64(id), 10), userFacingError(err))
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/rooms/"+strconv.FormatUint(uint64(id), 10))
}

func (h *AdminRoomHandler) renderForm(c *gin.Context, status int, room model.Room, action string, title string, errorMessage string) {
	room = roomFormDefaults(room)
	h.renderer.Render(c, status, "admin_base.html", "admin/room_form.html", gin.H{
		"Title":        title,
		"Room":         room,
		"Action":       action,
		"Statuses":     roomStatusOptions(),
		"RentTypes":    rentTypeOptions(),
		"PaymentTerms": paymentTermsOptions(),
		"Orientations": orientationOptions(),
		"Error":        errorMessage,
	})
}

func roomFormDefaults(room model.Room) model.Room {
	if room.RentType == "" {
		room.RentType = model.RentTypeMonthly
	}
	if room.PaymentTerms == "" {
		room.PaymentTerms = model.PaymentTerms1M1D
	}
	if room.RentPrice == 0 {
		room.RentPrice = room.Price
	}
	if room.Area == 0 {
		room.Area = 1
	}
	return room
}

func roomInputFromForm(c *gin.Context) (service.RoomInput, error) {
	input := service.RoomInput{
		RoomNo:        c.PostForm("room_no"),
		Title:         c.PostForm("title"),
		Description:   c.PostForm("description"),
		RentType:      c.PostForm("rent_type"),
		RentPriceYuan: c.PostForm("rent_price"),
		PriceYuan:     c.PostForm("price"),
		PaymentTerms:  c.PostForm("payment_terms"),
		DepositYuan:   c.PostForm("deposit"),
		Status:        c.PostForm("status"),
		Address:       c.PostForm("address"),
		Orientation:   c.PostForm("orientation"),
		Tags:          c.PostForm("tags"),
	}
	area, err := parseIntForm(c, "area")
	if err != nil {
		return input, err
	}
	input.Area = area
	floor, err := parseIntForm(c, "floor")
	if err != nil {
		return input, err
	}
	input.Floor = floor
	bedrooms, err := parseIntForm(c, "bedrooms")
	if err != nil {
		return input, err
	}
	input.Bedrooms = bedrooms
	livingRooms, err := parseIntForm(c, "living_rooms")
	if err != nil {
		return input, err
	}
	input.LivingRooms = livingRooms
	bathrooms, err := parseIntForm(c, "bathrooms")
	if err != nil {
		return input, err
	}
	input.Bathrooms = bathrooms
	return input, nil
}

func roomFromInput(input service.RoomInput) model.Room {
	rentPrice, _ := service.ParseIntegerYuanToFen(input.RentPriceYuan)
	if rentPrice == 0 {
		rentPrice, _ = service.ParseIntegerYuanToFen(input.PriceYuan)
	}
	deposit, _ := service.ParseIntegerYuanToFen(input.DepositYuan)
	return model.Room{
		RoomNo:       input.RoomNo,
		Title:        input.Title,
		Description:  input.Description,
		Price:        rentPrice,
		RentType:     model.RentTypeOrDefault(input.RentType),
		RentPrice:    rentPrice,
		PaymentTerms: model.PaymentTermsOrDefault(input.PaymentTerms),
		Deposit:      deposit,
		Status:       input.Status,
		Area:         input.Area,
		Floor:        input.Floor,
		Address:      input.Address,
		Bedrooms:     input.Bedrooms,
		LivingRooms:  input.LivingRooms,
		Bathrooms:    input.Bathrooms,
		Orientation:  input.Orientation,
		Tags:         input.Tags,
	}
}

func roomToAPIItem(room model.Room, tenant model.Tenant) roomAPIItem {
	item := roomAPIItem{
		ID:            room.ID,
		RoomNo:        room.RoomNo,
		Title:         room.Title,
		Status:        room.Status,
		StatusLabel:   roomStatusLabelText(room.Status),
		RentPriceFen:  model.RoomRentPrice(room),
		RentPriceText: service.FormatFen(model.RoomRentPrice(room)),
		RentType:      room.RentType,
		RentTypeLabel: rentTypeLabelText(room.RentType),
		Floor:         room.Floor,
		Area:          room.Area,
		DetailURL:     "/admin/rooms/" + strconv.FormatUint(uint64(room.ID), 10),
	}
	if tenant.ID > 0 {
		item.TenantID = tenant.ID
		item.TenantName = tenant.Name
		item.LeaseEndDate = formatOptionalAPIDate(tenant.LeaseEndDate)
	}
	return item
}

func roomStatusLabelText(status string) string {
	switch status {
	case model.RoomStatusVacant:
		return "空置"
	case model.RoomStatusOccupied:
		return "已出租"
	case model.RoomStatusMaintenance:
		return "维护中"
	default:
		return "未知"
	}
}

func roomStatusOptions() []SelectOption {
	return []SelectOption{
		{Value: model.RoomStatusVacant, Label: "空置"},
		{Value: model.RoomStatusOccupied, Label: "已出租"},
		{Value: model.RoomStatusMaintenance, Label: "维护中"},
	}
}

func rentTypeOptions() []SelectOption {
	return []SelectOption{
		{Value: model.RentTypeMonthly, Label: "月租（元）"},
		{Value: model.RentTypeDaily, Label: "日租（元）"},
	}
}

func paymentTermsOptions() []SelectOption {
	return []SelectOption{
		{Value: model.PaymentTerms1M1D, Label: "付一押一"},
		{Value: model.PaymentTerms1M2D, Label: "付一押二"},
		{Value: model.PaymentTerms3M1D, Label: "付三押一"},
		{Value: model.PaymentTerms6M0D, Label: "半年付"},
		{Value: model.PaymentTerms12M0D, Label: "年付"},
	}
}

func orientationOptions() []SelectOption {
	return []SelectOption{
		{Value: "东", Label: "东"},
		{Value: "南", Label: "南"},
		{Value: "西", Label: "西"},
		{Value: "北", Label: "北"},
		{Value: "东南", Label: "东南"},
		{Value: "东北", Label: "东北"},
		{Value: "西南", Label: "西南"},
		{Value: "西北", Label: "西北"},
		{Value: "南北", Label: "南北"},
		{Value: "东西", Label: "东西"},
	}
}
