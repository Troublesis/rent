package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/service"
)

type PublicHandler struct {
	renderer        Renderer
	roomService     *service.RoomService
	settingsService *service.SettingsService
}

func NewPublicHandler(renderer Renderer, roomService *service.RoomService, settingsService *service.SettingsService) *PublicHandler {
	return &PublicHandler{renderer: renderer, roomService: roomService, settingsService: settingsService}
}

func (h *PublicHandler) Index(c *gin.Context) {
	rooms, err := h.roomService.ListAvailableRooms(6, 0)
	if err != nil {
		c.String(http.StatusInternalServerError, "读取房源失败: %v", err)
		return
	}
	settings, err := h.settingsService.GetSettings()
	if err != nil {
		c.String(http.StatusInternalServerError, "读取联系信息失败: %v", err)
		return
	}
	h.renderer.Render(c, http.StatusOK, "public_base.html", "public/index.html", gin.H{
		"Title":    "安心租房",
		"Rooms":    rooms,
		"Settings": settings,
	})
}

func (h *PublicHandler) Rooms(c *gin.Context) {
	rooms, err := h.roomService.ListAvailableRooms(0, 0)
	if err != nil {
		c.String(http.StatusInternalServerError, "读取房源失败: %v", err)
		return
	}
	settings, err := h.settingsService.GetSettings()
	if err != nil {
		c.String(http.StatusInternalServerError, "读取联系信息失败: %v", err)
		return
	}
	h.renderer.Render(c, http.StatusOK, "public_base.html", "public/rooms.html", gin.H{
		"Title":    "可租房源",
		"Rooms":    rooms,
		"Settings": settings,
	})
}

func (h *PublicHandler) RoomDetail(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	room, err := h.roomService.GetRoom(id)
	if err != nil || room.Status != model.RoomStatusVacant {
		c.String(http.StatusNotFound, "房源不存在或暂不可租")
		return
	}
	settings, err := h.settingsService.GetSettings()
	if err != nil {
		c.String(http.StatusInternalServerError, "读取联系信息失败: %v", err)
		return
	}
	h.renderer.Render(c, http.StatusOK, "public_base.html", "public/room_detail.html", gin.H{
		"Title":    room.Title,
		"Room":     room,
		"Settings": settings,
	})
}
