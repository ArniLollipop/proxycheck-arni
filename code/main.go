package main

import (
	"context"
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

func main() {
	// Initialize a single database
	db, err := gorm.Open(sqlite.Open("database/proxy.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	var wg sync.WaitGroup
	quit := make(chan struct{})

	// Auto-migrate all models
	if err := db.AutoMigrate(&Proxy{}, &Settings{}, &ProxySpeedLog{}, &ProxyIPLog{}, &ProxyVisitLogs{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Initialize Settings from the single database
	settings := SettingsDefault(db)

	// Initialize Gin router
	router := gin.Default()

	// Serve frontend static files
	router.Use(static.Serve("/", static.LocalFile("./client/dist", true)))

	// Init Geoip service
	geoIP, err := NewGeoIPClient("GeoIP2-ISP.mmdb")
	if err != nil {
		log.Fatalf("failed to initialize GeoIP service: %v", err)
	}

	// Запускаем планировщик проверки IP в отдельной горутине.
	go StartIPCheckScheduler(&wg, quit, db, settings, geoIP)
	go StartHealthCheckScheduler(&wg, quit, db, settings)

	// Create handler instance
	h := handler{
		db:          db,
		settings:    settings,
		geoIPClient: geoIP,
	}

	// API routes for proxies
	proxyRoutes := router.Group("api/proxy")
	{
		proxyRoutes.GET("", h.ProxyList)
		proxyRoutes.PUT(":id", h.UpdateProxy)
		proxyRoutes.POST("", h.CreateProxy)
		proxyRoutes.GET(":id/verify", h.Verify)
		proxyRoutes.DELETE(":id", h.Delete)
		proxyRoutes.POST("verify-batch", h.VerifyBatch)
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
	router.POST("api/import", h.ImportProxies)
	router.GET("/api/speedLogs", h.GetSpeedLogs)
	router.GET("/api/ipLogs", h.GetProxyIPLogs)
	router.POST("/api/proxyVisits", h.CreateProxyVisitLog)
	router.GET("/api/proxyVisits", h.GetProxyVisitLogs)

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
	<-stop // Ожидаем сигнала завершения.
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	close(quit)
	wg.Wait() // Ожидаем завершения всех горутин.

	log.Println("Server exiting")
}

//Speed - почему килабити
// Пароль - спраятать
// IP - local ip
// Import - точно также как и экспорт
// Дати не работают
