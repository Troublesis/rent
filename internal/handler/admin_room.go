package handler

import (
	"net/http"
	"net/url"
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

type adminRoomStatusLink struct {
	Label  string
	URL    string
	Active bool
}

type adminRoomSortLink struct {
	Label     string
	URL       string
	Active    bool
	Indicator string
}

func NewAdminRoomHandler(renderer Renderer, roomService *service.RoomService, tenantService *service.TenantService) *AdminRoomHandler {
	return &AdminRoomHandler{renderer: renderer, roomService: roomService, tenantService: tenantService}
}

const (
	adminRoomStatusAll     = "all"
	adminRoomViewList      = "list"
	adminRoomViewCard      = "card"
	adminRoomSortRoomNo    = "room_no"
	adminRoomSortTitle     = "title"
	adminRoomSortRentPrice = "rent_price"
	adminRoomSortStatus    = "status"
	adminRoomSortAsc       = "asc"
	adminRoomSortDesc      = "desc"
)

func adminRoomViewFromQuery(c *gin.Context) string {
	if c.Query("view") == adminRoomViewCard {
		return adminRoomViewCard
	}
	return adminRoomViewList
}

func adminRoomStatusFromQuery(c *gin.Context) (string, string) {
	switch c.Query("status") {
	case adminRoomStatusAll:
		return "", adminRoomStatusAll
	case model.RoomStatusOccupied:
		return model.RoomStatusOccupied, model.RoomStatusOccupied
	case model.RoomStatusMaintenance:
		return model.RoomStatusMaintenance, model.RoomStatusMaintenance
	default:
		return model.RoomStatusVacant, model.RoomStatusVacant
	}
}

func adminRoomSortFromQuery(c *gin.Context) (string, string) {
	sortBy := adminRoomSortRoomNo
	switch c.Query("sort_by") {
	case adminRoomSortTitle:
		sortBy = adminRoomSortTitle
	case adminRoomSortRentPrice:
		sortBy = adminRoomSortRentPrice
	case adminRoomSortStatus:
		sortBy = adminRoomSortStatus
	}
	sortDir := adminRoomSortAsc
	if c.Query("sort_dir") == adminRoomSortDesc {
		sortDir = adminRoomSortDesc
	}
	return sortBy, sortDir
}

func adminRoomsURLWithStatus(queryText string, status string, viewMode string, sortBy string, sortDir string) string {
	values := url.Values{}
	if queryText != "" {
		values.Set("q", queryText)
	}
	if status != "" {
		values.Set("status", status)
	}
	if viewMode != "" {
		values.Set("view", viewMode)
	}
	if sortBy != "" {
		values.Set("sort_by", sortBy)
	}
	if sortDir != "" {
		values.Set("sort_dir", sortDir)
	}
	query := values.Encode()
	if query == "" {
		return "/admin/rooms"
	}
	return "/admin/rooms?" + query
}

func adminRoomsURL(filter repository.RoomFilter, viewMode string) string {
	return adminRoomsURLWithStatus(filter.Query, filter.Status, viewMode, filter.SortBy, filter.SortDir)
}

func adminRoomStatusLinks(queryText string, currentStatus string, viewMode string, sortBy string, sortDir string) []adminRoomStatusLink {
	options := []SelectOption{
		{Value: model.RoomStatusVacant, Label: "空置"},
		{Value: model.RoomStatusOccupied, Label: "已出租"},
		{Value: model.RoomStatusMaintenance, Label: "维护中"},
		{Value: adminRoomStatusAll, Label: "全部"},
	}
	links := make([]adminRoomStatusLink, len(options))
	for i, option := range options {
		links[i] = adminRoomStatusLink{
			Label:  option.Label,
			URL:    adminRoomsURLWithStatus(queryText, option.Value, viewMode, sortBy, sortDir),
			Active: currentStatus == option.Value,
		}
	}
	return links
}

func adminRoomSortLinks(queryText string, currentStatus string, viewMode string, sortBy string, sortDir string) []adminRoomSortLink {
	options := []SelectOption{
		{Value: adminRoomSortRoomNo, Label: "房间号"},
		{Value: adminRoomSortTitle, Label: "标题"},
		{Value: adminRoomSortRentPrice, Label: "租金金额"},
		{Value: adminRoomSortStatus, Label: "状态"},
	}
	links := make([]adminRoomSortLink, len(options))
	for i, option := range options {
		nextDir := adminRoomSortAsc
		active := sortBy == option.Value
		if active && sortDir == adminRoomSortAsc {
			nextDir = adminRoomSortDesc
		}
		links[i] = adminRoomSortLink{
			Label:     option.Label,
			URL:       adminRoomsURLWithStatus(queryText, currentStatus, viewMode, option.Value, nextDir),
			Active:    active,
			Indicator: adminRoomSortIndicator(active, sortDir),
		}
	}
	return links
}

func adminRoomSortIndicator(active bool, sortDir string) string {
	if !active {
		return ""
	}
	if sortDir == adminRoomSortDesc {
		return "↓"
	}
	return "↑"
}

func adminRoomCurrentStatusLabel(currentStatus string) string {
	if currentStatus == adminRoomStatusAll {
		return "全部"
	}
	return roomStatusLabelText(currentStatus)
}

func adminRoomBaseURL(id uint) string {
	return "/admin/rooms/" + strconv.FormatUint(uint64(id), 10)
}

func adminRoomEditURL(id uint) string {
	return adminRoomBaseURL(id) + "/edit"
}

func (h *AdminRoomHandler) List(c *gin.Context) {
	filterStatus, currentStatus := adminRoomStatusFromQuery(c)
	sortBy, sortDir := adminRoomSortFromQuery(c)
	filter := repository.RoomFilter{Status: filterStatus, Query: c.Query("q"), SortBy: sortBy, SortDir: sortDir}
	viewMode := adminRoomViewFromQuery(c)
	rooms, err := h.roomService.ListRooms(filter)
	if err != nil {
		c.String(http.StatusInternalServerError, "读取房源失败")
		return
	}
	h.renderer.Render(c, http.StatusOK, "admin_base.html", "admin/rooms.html", gin.H{
		"Title":              "房源管理",
		"Rooms":              rooms,
		"Statuses":           roomStatusOptions(),
		"StatusLinks":        adminRoomStatusLinks(filter.Query, currentStatus, viewMode, sortBy, sortDir),
		"SortLinks":          adminRoomSortLinks(filter.Query, currentStatus, viewMode, sortBy, sortDir),
		"CurrentStatus":      currentStatus,
		"CurrentStatusLabel": adminRoomCurrentStatusLabel(currentStatus),
		"SortBy":             sortBy,
		"SortDir":            sortDir,
		"Filter":             filter,
		"ViewMode":           viewMode,
		"ListViewURL":        adminRoomsURLWithStatus(filter.Query, currentStatus, adminRoomViewList, sortBy, sortDir),
		"CardViewURL":        adminRoomsURLWithStatus(filter.Query, currentStatus, adminRoomViewCard, sortBy, sortDir),
		"ClearFilterURL":     adminRoomsURLWithStatus("", model.RoomStatusVacant, viewMode, adminRoomSortRoomNo, adminRoomSortAsc),
		"Error":              queryError(c),
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
	h.renderForm(c, http.StatusOK, room, nil, "/admin/rooms", "新增房源", "")
}

func (h *AdminRoomHandler) Create(c *gin.Context) {
	input, err := roomInputFromForm(c)
	if err != nil {
		h.renderForm(c, http.StatusBadRequest, roomFromInput(input), nil, "/admin/rooms", "新增房源", "表单数据不正确")
		return
	}
	if _, err := h.roomService.CreateRoom(input); err != nil {
		h.renderForm(c, http.StatusBadRequest, roomFromInput(input), nil, "/admin/rooms", "新增房源", userFacingError(err))
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/rooms")
}

func (h *AdminRoomHandler) Detail(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	c.Redirect(http.StatusSeeOther, adminRoomEditURL(id))
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
	tenants := h.roomTenantHistory(id)
	h.renderForm(c, http.StatusOK, *room, tenants, adminRoomBaseURL(id), "编辑房源", "")
}

func (h *AdminRoomHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	input, err := roomInputFromForm(c)
	if err != nil {
		h.renderForm(c, http.StatusBadRequest, h.roomFromFailedUpdate(id, input), h.roomTenantHistory(id), adminRoomBaseURL(id), "编辑房源", "表单数据不正确")
		return
	}
	if _, err := h.roomService.UpdateRoom(id, input); err != nil {
		h.renderForm(c, http.StatusBadRequest, h.roomFromFailedUpdate(id, input), h.roomTenantHistory(id), adminRoomBaseURL(id), "编辑房源", userFacingError(err))
		return
	}
	c.Redirect(http.StatusSeeOther, adminRoomEditURL(id))
}

func (h *AdminRoomHandler) roomTenantHistory(roomID uint) []model.Tenant {
	if h.tenantService == nil {
		return nil
	}
	tenants, err := h.tenantService.ListTenantsByRoomID(roomID)
	if err != nil {
		return nil
	}
	return tenants
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
		redirectWithError(c, adminRoomEditURL(id), userFacingError(err))
		return
	}
	c.Redirect(http.StatusSeeOther, adminRoomEditURL(id))
}

func (h *AdminRoomHandler) roomFromFailedUpdate(id uint, input service.RoomInput) model.Room {
	room := roomFromInput(input)
	room.ID = id
	currentRoom, err := h.roomService.GetRoom(id)
	if err == nil {
		room.Media = currentRoom.Media
	}
	return room
}

func (h *AdminRoomHandler) renderForm(c *gin.Context, status int, room model.Room, tenants []model.Tenant, action string, title string, errorMessage string) {
	room = roomFormDefaults(room)
	h.renderer.Render(c, status, "admin_base.html", "admin/room_form.html", gin.H{
		"Title":        title,
		"Room":         room,
		"Tenants":      tenants,
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
		RentPriceText: service.FormatFenAsYuanInt(model.RoomRentPrice(room)),
		RentType:      room.RentType,
		RentTypeLabel: rentTypeLabelText(room.RentType),
		Floor:         room.Floor,
		Area:          room.Area,
		DetailURL:     adminRoomEditURL(room.ID),
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
