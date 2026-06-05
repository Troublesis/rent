package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/troublesis/rent/internal/service"
)

type AdminPushHandler struct {
	renderer    Renderer
	pushService *service.PushService
}

func NewAdminPushHandler(renderer Renderer, pushService *service.PushService) *AdminPushHandler {
	return &AdminPushHandler{renderer: renderer, pushService: pushService}
}

func (h *AdminPushHandler) UpdatePushSettings(c *gin.Context) {
	pushMode := strings.TrimSpace(c.PostForm("push_mode"))
	pushTime := strings.TrimSpace(c.PostForm("push_time"))
	scheduleMode := strings.TrimSpace(c.PostForm("schedule_mode"))
	advanceDaysStr := strings.TrimSpace(c.PostForm("advance_days"))
	fixedPeriod := strings.TrimSpace(c.PostForm("fixed_period"))
	fixedWeekdayStr := strings.TrimSpace(c.PostForm("fixed_weekday"))
	fixedMonthDayStr := strings.TrimSpace(c.PostForm("fixed_month_day"))
	token := strings.TrimSpace(c.PostForm("pushplus_token"))
	enabledStr := strings.TrimSpace(c.PostForm("push_enabled"))
	titleTpl := strings.TrimSpace(c.PostForm("template_title"))
	bodyTpl := strings.TrimSpace(c.PostForm("template_body"))

	config, err := h.pushService.GetConfig()
	if err != nil {
		c.String(http.StatusInternalServerError, "读取推送配置失败")
		return
	}

	if pushMode != "" {
		config.PushMode = pushMode
	}
	if pushTime != "" {
		config.PushTime = pushTime
	}
	if scheduleMode != "" {
		config.ScheduleMode = scheduleMode
	}
	if scheduleMode == "advance" && advanceDaysStr != "" {
		fmt.Sscanf(advanceDaysStr, "%d", &config.AdvanceDays)
	}
	if scheduleMode == "fixed" {
		config.FixedPeriod = fixedPeriod
		if fixedPeriod == "weekly" {
			fmt.Sscanf(fixedWeekdayStr, "%d", &config.FixedWeekday)
		}
		if fixedPeriod == "monthly" {
			fmt.Sscanf(fixedMonthDayStr, "%d", &config.FixedMonthDay)
		}
	}
	config.PushPlusToken = token
	config.Enabled = (enabledStr == "on" || enabledStr == "true" || enabledStr == "1")

	if err := h.pushService.SaveConfig(config); err != nil {
		c.String(http.StatusInternalServerError, "保存推送配置失败")
		return
	}

	tmpl, err := h.pushService.GetTemplate()
	if err != nil {
		c.String(http.StatusInternalServerError, "读取消息模板失败")
		return
	}
	if titleTpl != "" {
		tmpl.TitleTpl = titleTpl
	}
	if bodyTpl != "" {
		tmpl.BodyTpl = bodyTpl
	}
	if err := h.pushService.SaveTemplate(tmpl); err != nil {
		c.String(http.StatusInternalServerError, "保存消息模板失败")
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/settings")
}

func (h *AdminPushHandler) TestPush(c *gin.Context) {
	token := strings.TrimSpace(c.PostForm("pushplus_token"))
	title := strings.TrimSpace(c.PostForm("template_title"))
	body := strings.TrimSpace(c.PostForm("template_body"))

	if err := h.pushService.SendTest(token, title, body); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "测试消息已发送，请查看微信"})
}
