package main

import (
	"fmt"
	"log"

	"radius-buildup/internal/config"
	"radius-buildup/internal/database"
	"radius-buildup/internal/middleware"
	"radius-buildup/internal/modules/activity"
	"radius-buildup/internal/modules/admin"
	"radius-buildup/internal/modules/auth"
	"radius-buildup/internal/modules/billing"
	"radius-buildup/internal/modules/company"
	"radius-buildup/internal/modules/cron"
	"radius-buildup/internal/modules/customer"
	"radius-buildup/internal/modules/dashboard"
	"radius-buildup/internal/modules/health"
	"radius-buildup/internal/modules/hotspot"
	"radius-buildup/internal/modules/mikrotik"
	"radius-buildup/internal/modules/olt"
	"radius-buildup/internal/modules/payment"
	"radius-buildup/internal/modules/pppoe"
	"radius-buildup/internal/modules/router"
	"radius-buildup/internal/modules/sessions"
	"radius-buildup/internal/modules/whatsapp"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	log.Printf("🚀 Starting Radius Buildup Go API [%s]", cfg.Env)

	// Connect to database
	_, err = database.Connect(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Initialize background cron jobs
	cron.InitCron()
	defer cron.StopCron()

	// Setup Gin
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Middleware
	r.Use(middleware.CORS())

	// API Routes
	api := r.Group("/api")
	{
		// Public routes (no auth needed)
		health.RegisterRoutes(api)
		auth.RegisterRoutes(api, cfg)
		company.RegisterPublicRoutes(api)
		payment.RegisterPublicRoutes(api)

		// Protected routes (JWT required)
		protected := api.Group("")
		protected.Use(middleware.AuthRequired(cfg))
		{
			auth.RegisterProtectedRoutes(protected)
			pppoe.RegisterRoutes(protected)
			router.RegisterRoutes(protected)
			billing.RegisterRoutes(protected)
			dashboard.RegisterRoutes(protected)
			sessions.RegisterRoutes(protected)
			company.RegisterRoutes(protected)
			activity.RegisterRoutes(protected)
			hotspot.RegisterRoutes(protected)
			admin.RegisterRoutes(protected)
			customer.RegisterRoutes(protected)
			mikrotik.RegisterRoutes(protected)
			payment.RegisterRoutes(protected)
			whatsapp.RegisterRoutes(protected)
			olt.RegisterRoutes(protected)
		}
	}

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":    cfg.AppName,
			"version": "1.0.0",
			"engine":  "Go + Gin",
			"docs":    fmt.Sprintf("http://localhost:%s/api/health", cfg.Port),
		})
	})

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🌐 Server listening on http://localhost%s", addr)
	log.Printf("📡 Health check: http://localhost%s/api/health", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
