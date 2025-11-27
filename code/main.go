package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NoBufferMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasSuffix(c.Request.URL.Path, "/verify-batch") {
			c.Writer.Header().Set("X-Accel-Buffering", "no")
		}
		c.Next()
	}
}

func main() {
	log.Println("Starting Proxy Checker application...")

	// Initialize a single database
	db, err := gorm.Open(sqlite.Open("database/proxy.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	var wg sync.WaitGroup
	quit := make(chan struct{})

	// Auto-migrate all models
	if err := db.AutoMigrate(&Proxy{}, &Settings{}, &ProxySpeedLog{}, &ProxyIPLog{}, &ProxyVisitLogs{}, &ProxyFailureLog{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Initialize Settings from the single database
	settings := SettingsDefault(db)

	// Initialize Notification Service
	notificationService := NewNotificationService(
		settings.TelegramEnabled,
		settings.TelegramToken,
		settings.TelegramChatID,
	)
	log.Printf("Notification service initialized. Telegram enabled: %v", settings.TelegramEnabled)

	// Канал для инициирования перезапуска из API
	restartSignal := make(chan struct{}, 1)

	// Initialize Gin router
	router := gin.Default()
	router.Use(NoBufferMiddleware())


	// Serve frontend static files
	router.Use(gin.BasicAuth(gin.Accounts{settings.Username: settings.Password}), static.Serve("/", static.LocalFile("./client/dist", true)))

	// Init Geoip service
	geoIP, err := NewGeoIPClient("GeoIP2-ISP.mmdb")
	if err != nil {
		log.Fatalf("failed to initialize GeoIP service: %v", err)
	}

	// Запускаем планировщик проверки IP в отдельной горутине.
	go StartIPCheckScheduler(&wg, quit, db, settings, geoIP, notificationService)
	go StartHealthCheckScheduler(&wg, quit, db, settings, notificationService)

	// Create handler instance
	h := handler{
		db:            db,
		settings:      settings,
		geoIPClient:   geoIP,
		restartSignal: restartSignal, // Передаем канал в обработчик
	}

	// API routes for proxies
	sseRoutes := router.Group("api/proxy")
	sseRoutes.Use(func(c *gin.Context) {
			// Отключаем логирование и другие middleware для SSE
			c.Next()
	})
	sseRoutes.GET("verify-batch", h.VerifyBatch)

	proxyRoutes := router.Group("api/proxy")
	{ 
		proxyRoutes.GET("", h.ProxyList)
    proxyRoutes.PUT(":id", h.UpdateProxy)
    proxyRoutes.POST("", h.CreateProxy)
    proxyRoutes.GET(":id/verify", h.Verify)
    proxyRoutes.DELETE(":id", h.Delete)
	}



	// API routes for settings

	settingsRoutes := router.Group("api/settings")

	{
		settingsRoutes.GET("", h.GetSettings)
		settingsRoutes.PUT("", h.UpdateSettings)
	}

	// Export routes
	exportRoutes := router.Group("api/export")

	{
		exportRoutes.GET("all", h.ExportAll)
		exportRoutes.GET("selected", h.ExportSelected)
	}
	
	router.POST("api/import", func(c *gin.Context) {
		// Import proxies and run checks if successful
		if err := h.ImportProxies(c); err != nil {
			// Error response is already handled in ImportProxies
			return
		}

		go RunSingleIPCheck(db, settings, geoIP, notificationService)
		go RunSingleHealthCheck(db, settings, notificationService)
	})
	router.GET("/api/speedLogs", h.GetSpeedLogs)
	router.GET("/api/ipLogs", h.GetProxyIPLogs)
	router.POST("/api/proxyVisits", h.CreateProxyVisitLog)
	router.GET("/api/proxyVisits", h.GetProxyVisitLogs)
	router.GET("/api/failureLogs", h.GetFailureLogs)
	router.GET("/api/failureStats/:id", h.GetFailureStats)
	router.POST("/api/testNotification", h.TestNotification)

	// Handle SPA routing (Vue Router history mode)
	router.NoRoute(func(c *gin.Context) {
		// Return index.html for any non-api route
		if !strings.HasPrefix(c.Request.RequestURI, "/api") {
			c.File("./client/dist/index.html")
		}
	})

	// Setup HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Run server in a goroutine
	go func() {
		log.Println("Server running on http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Ожидаем сигнала завершения (либо от ОС, либо от нашего API)
	select {
	case <-stop:
		log.Println("Received shutdown signal from OS.")
	case <-restartSignal:
		log.Println("Received restart signal from API.")
	}

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	close(quit)
	wg.Wait() // Ожидаем завершения всех горутин.

	// Cleanup GeoIP client
	if err := geoIP.Close(); err != nil {
		log.Printf("Error closing GeoIP client: %v", err)
	}

	log.Println("Server exiting")
}
