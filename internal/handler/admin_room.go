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
	renderer    Renderer
	roomService *service.RoomService
}

func NewAdminRoomHandler(renderer Renderer, roomService *service.RoomService) *AdminRoomHandler {
	return &AdminRoomHandler{renderer: renderer, roomService: roomService}
}

func (h *AdminRoomHandler) List(c *gin.Context) {
	filter := repository.RoomFilter{Status: c.Query("status"), Query: c.Query("q")}
	rooms, err := h.roomService.ListRooms(filter)
	if err != nil {
		c.String(http.StatusInternalServerError, "读取房源失败: %v", err)
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

func (h *AdminRoomHandler) New(c *gin.Context) {
	h.renderForm(c, http.StatusOK, model.Room{Status: model.RoomStatusVacant}, "/admin/rooms", "新增房源", "")
}

func (h *AdminRoomHandler) Create(c *gin.Context) {
	input, err := roomInputFromForm(c)
	if err != nil {
		h.renderForm(c, http.StatusBadRequest, roomFromInput(input), "/admin/rooms", "新增房源", "表单数据不正确")
		return
	}
	if _, err := h.roomService.CreateRoom(input); err != nil {
		h.renderForm(c, http.StatusBadRequest, roomFromInput(input), "/admin/rooms", "新增房源", err.Error())
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
		h.renderForm(c, http.StatusBadRequest, roomFromInput(input), "/admin/rooms/"+strconv.FormatUint(uint64(id), 10), "编辑房源", err.Error())
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
		redirectWithError(c, "/admin/rooms", err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/rooms")
}

func (h *AdminRoomHandler) renderForm(c *gin.Context, status int, room model.Room, action string, title string, errorMessage string) {
	h.renderer.Render(c, status, "admin_base.html", "admin/room_form.html", gin.H{
		"Title":    title,
		"Room":     room,
		"Action":   action,
		"Statuses": roomStatusOptions(),
		"Error":    errorMessage,
	})
}

func roomInputFromForm(c *gin.Context) (service.RoomInput, error) {
	area, err := parseIntForm(c, "area")
	if err != nil {
		return service.RoomInput{}, err
	}
	floor, err := parseIntForm(c, "floor")
	if err != nil {
		return service.RoomInput{}, err
	}
	return service.RoomInput{
		RoomNo:      c.PostForm("room_no"),
		Title:       c.PostForm("title"),
		Description: c.PostForm("description"),
		PriceYuan:   c.PostForm("price"),
		DepositYuan: c.PostForm("deposit"),
		Status:      c.PostForm("status"),
		Area:        area,
		Floor:       floor,
		Tags:        c.PostForm("tags"),
	}, nil
}

func roomFromInput(input service.RoomInput) model.Room {
	price, _ := service.ParseYuanToFen(input.PriceYuan)
	deposit, _ := service.ParseYuanToFen(input.DepositYuan)
	return model.Room{
		RoomNo:      input.RoomNo,
		Title:       input.Title,
		Description: input.Description,
		Price:       price,
		Deposit:     deposit,
		Status:      input.Status,
		Area:        input.Area,
		Floor:       input.Floor,
		Tags:        input.Tags,
	}
}

func roomStatusOptions() []SelectOption {
	return []SelectOption{
		{Value: model.RoomStatusVacant, Label: "空置"},
		{Value: model.RoomStatusOccupied, Label: "已出租"},
		{Value: model.RoomStatusMaintenance, Label: "维护中"},
	}
}
