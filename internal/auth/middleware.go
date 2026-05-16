package auth

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const SessionAdminLoggedIn = "admin_logged_in"

func RequireLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		loggedIn, ok := session.Get(SessionAdminLoggedIn).(bool)
		if ok && loggedIn {
			c.Next()
			return
		}
		if c.Request.Header.Get("HX-Request") == "true" {
			c.Header("HX-Redirect", "/admin/login")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Redirect(http.StatusSeeOther, "/admin/login")
		c.Abort()
	}
}
