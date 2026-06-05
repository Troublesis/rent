package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/service"
)

type AdminSettingsHandler struct {
	renderer        Renderer
	settingsService *service.SettingsService
}

func NewAdminSettingsHandler(renderer Renderer, settingsService *service.SettingsService) *AdminSettingsHandler {
	return &AdminSettingsHandler{renderer: renderer, settingsService: settingsService}
}

func (h *AdminSettingsHandler) Page(c *gin.Context) {
	settings, err := h.settingsService.GetSettings()
	if err != nil {
		c.String(http.StatusInternalServerError, "读取设置失败")
		return
	}
	pushConfig, _ := h.settingsService.GetPushConfig()
	pushTemplate, _ := h.settingsService.GetPushTemplate()
	h.render(c, http.StatusOK, settings, pushConfig, pushTemplate, queryError(c))
}

func (h *AdminSettingsHandler) Update(c *gin.Context) {
	settings := service.Settings{
		LandlordName:  strings.TrimSpace(c.PostForm("landlord_name")),
		LandlordPhone: strings.TrimSpace(c.PostForm("landlord_phone")),
	}
	if settings.LandlordName == "" || settings.LandlordPhone == "" {
		pushConfig, _ := h.settingsService.GetPushConfig()
		pushTemplate, _ := h.settingsService.GetPushTemplate()
		h.render(c, http.StatusBadRequest, settings, pushConfig, pushTemplate, "房东姓名和联系电话不能为空")
		return
	}
	if err := h.settingsService.UpdateSettings(settings); err != nil {
		pushConfig, _ := h.settingsService.GetPushConfig()
		pushTemplate, _ := h.settingsService.GetPushTemplate()
		h.render(c, http.StatusInternalServerError, settings, pushConfig, pushTemplate, userFacingError(err))
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/settings")
}

func (h *AdminSettingsHandler) render(c *gin.Context, status int, settings service.Settings, pushConfig *model.PushConfig, pushTemplate *model.PushTemplate, errorMessage string) {
	h.renderer.Render(c, status, "admin_base.html", "admin/settings.html", gin.H{
		"Title":        "系统设置",
		"Settings":     settings,
		"PushConfig":   pushConfig,
		"PushTemplate": pushTemplate,
		"Error":        errorMessage,
	})
}
