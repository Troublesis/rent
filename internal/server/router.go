package server

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/troublesis/rent/config"
	"github.com/troublesis/rent/internal/auth"
	"github.com/troublesis/rent/internal/handler"
	"github.com/troublesis/rent/internal/repository"
	"github.com/troublesis/rent/internal/service"
	"gorm.io/gorm"
)

func NewRouter(cfg config.Config, db *gorm.DB) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.MaxMultipartMemory = 10 << 20
	router.Static("/static", "./static")
	router.Static("/uploads", cfg.UploadDir)

	store := cookie.NewStore([]byte(cfg.SessionSecret))
	store.Options(sessions.Options{Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
	router.Use(sessions.Sessions("rent_session", store))

	renderer := NewTemplateRenderer("templates")
	roomRepo := repository.NewRoomRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)

	roomService := service.NewRoomService(roomRepo, tenantRepo)
	tenantService := service.NewTenantService(db, tenantRepo, roomRepo)
	paymentService := service.NewPaymentService(paymentRepo, tenantRepo)
	dashboardService := service.NewDashboardService(roomRepo, tenantRepo, paymentRepo)
	settingsService := service.NewSettingsService(cfg, settingsRepo)

	publicHandler := handler.NewPublicHandler(renderer, roomService, settingsService)
	authHandler := handler.NewAuthHandler(renderer, cfg)
	dashboardHandler := handler.NewAdminDashboardHandler(renderer, dashboardService)
	roomHandler := handler.NewAdminRoomHandler(renderer, roomService)
	tenantHandler := handler.NewAdminTenantHandler(renderer, tenantService, roomService)
	paymentHandler := handler.NewAdminPaymentHandler(renderer, paymentService, tenantService)
	statsHandler := handler.NewAdminStatsHandler(renderer, paymentService)
	settingsHandler := handler.NewAdminSettingsHandler(renderer, settingsService)
	uploadHandler := handler.NewUploadHandler(cfg.UploadDir, roomService)

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/", publicHandler.Index)
	router.GET("/rooms", publicHandler.Rooms)
	router.GET("/room/:id", publicHandler.RoomDetail)

	router.GET("/admin/login", authHandler.LoginPage)
	router.POST("/admin/login", authHandler.Login)
	router.POST("/admin/logout", authHandler.Logout)

	admin := router.Group("/admin")
	admin.Use(auth.RequireLogin())
	admin.GET("/dashboard", dashboardHandler.Dashboard)
	admin.GET("/rooms", roomHandler.List)
	admin.GET("/rooms/new", roomHandler.New)
	admin.POST("/rooms", roomHandler.Create)
	admin.GET("/rooms/:id", roomHandler.Detail)
	admin.GET("/rooms/:id/edit", roomHandler.Edit)
	admin.POST("/rooms/:id", roomHandler.Update)
	admin.POST("/rooms/:id/delete", roomHandler.Delete)
	admin.GET("/tenants", tenantHandler.List)
	admin.GET("/tenants/new", tenantHandler.New)
	admin.POST("/tenants/checkin", tenantHandler.CheckIn)
	admin.GET("/tenants/:id", tenantHandler.Detail)
	admin.POST("/tenants/:id/checkout", tenantHandler.CheckOut)
	admin.GET("/payments", paymentHandler.List)
	admin.POST("/payments", paymentHandler.Create)
	admin.POST("/payments/:id/toggle", paymentHandler.Toggle)
	admin.GET("/stats", statsHandler.Page)
	admin.GET("/settings", settingsHandler.Page)
	admin.POST("/settings", settingsHandler.Update)

	api := router.Group("/api")
	api.Use(auth.RequireLogin())
	api.POST("/upload", uploadHandler.UploadRoomMedia)
	api.GET("/dashboard/stats", statsHandler.DashboardStats)

	return router
}
