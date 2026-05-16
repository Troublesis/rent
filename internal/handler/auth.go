package handler

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/troublesis/rent/config"
	"github.com/troublesis/rent/internal/auth"
)

type AuthHandler struct {
	renderer Renderer
	cfg      config.Config
}

func NewAuthHandler(renderer Renderer, cfg config.Config) *AuthHandler {
	return &AuthHandler{renderer: renderer, cfg: cfg}
}

func (h *AuthHandler) LoginPage(c *gin.Context) {
	h.renderer.Render(c, http.StatusOK, "auth_base.html", "auth/login.html", gin.H{
		"Title": "房东登录",
		"Error": queryError(c),
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")
	if username != h.cfg.AdminUsername || password != h.cfg.AdminPassword {
		h.renderer.Render(c, http.StatusUnauthorized, "auth_base.html", "auth/login.html", gin.H{
			"Title":    "房东登录",
			"Error":    "账号或密码不正确",
			"Username": username,
		})
		return
	}

	session := sessions.Default(c)
	session.Set(auth.SessionAdminLoggedIn, true)
	session.Options(sessions.Options{Path: "/", MaxAge: 60 * 60 * 24 * 365 * 10, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	if err := session.Save(); err != nil {
		h.renderer.Render(c, http.StatusInternalServerError, "auth_base.html", "auth/login.html", gin.H{
			"Title": "房东登录",
			"Error": "登录状态保存失败，请重试",
		})
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/dashboard")
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	_ = session.Save()
	c.Redirect(http.StatusSeeOther, "/admin/login")
}
