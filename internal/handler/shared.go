package handler

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Renderer interface {
	Render(c *gin.Context, status int, layout string, page string, data gin.H)
	RenderPartial(c *gin.Context, status int, page string, templateName string, data gin.H)
}

type SelectOption struct {
	Value string
	Label string
}

func parseUintParam(c *gin.Context, name string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || value == 0 {
		c.String(http.StatusNotFound, "未找到资源")
		return 0, false
	}
	return uint(value), true
}

func parseUintForm(c *gin.Context, name string) (uint, error) {
	value, err := strconv.ParseUint(strings.TrimSpace(c.PostForm(name)), 10, 64)
	if err != nil || value == 0 {
		return 0, err
	}
	return uint(value), nil
}

func parseIntForm(c *gin.Context, name string) (int, error) {
	value := strings.TrimSpace(c.PostForm(name))
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

func parseDateForm(c *gin.Context, name string) (time.Time, error) {
	value := strings.TrimSpace(c.PostForm(name))
	if value == "" {
		return time.Time{}, nil
	}
	return time.ParseInLocation("2006-01-02", value, time.Local)
}

func redirectWithError(c *gin.Context, target string, message string) {
	separator := "?"
	if strings.Contains(target, "?") {
		separator = "&"
	}
	c.Redirect(http.StatusSeeOther, target+separator+"error="+url.QueryEscape(message))
}

func queryError(c *gin.Context) string {
	return strings.TrimSpace(c.Query("error"))
}
