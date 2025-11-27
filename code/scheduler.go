package main

import (
	"context"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

var healthMu sync.Mutex
var ipCheckMu sync.Mutex

const (
	// Number of concurrent workers for proxy checking
	MaxConcurrentWorkers = 10
)

func RunSingleIPCheck(db *gorm.DB, settings *Settings, geoIP *GeoIPClient, notifier *NotificationService) {
	var proxies []Proxy
	if err := db.Find(&proxies).Error; err != nil {
		log.Println("Error fetching proxies:", err)
		return
	}
	IPCheckIteratorWithNotifications(context.Background(), proxies, settings, db, geoIP, notifier)
}

func RunSingleHealthCheck(db *gorm.DB, settings *Settings, notifier *NotificationService) {
	var proxies []Proxy
	if err := db.Find(&proxies).Error; err != nil {
		log.Println("Error fetching proxies:", err)
		return
	}
	HealthCheckIteratorWithNotifications(context.Background(), proxies, settings, db, notifier)
}


func IPCheckIterator(proxies []Proxy, settings *Settings, db *gorm.DB, geoIPClient *GeoIPClient) {
	ctx := context.Background()
	IPCheckIteratorWithContext(ctx, proxies, settings, db, geoIPClient)
}

func IPCheckIteratorWithContext(ctx context.Context, proxies []Proxy, settings *Settings, db *gorm.DB, geoIPClient *GeoIPClient) {
	// Create a worker pool for concurrent checking
	var wg sync.WaitGroup
	proxyChan := make(chan *Proxy, len(proxies))

	// Start worker goroutines
	for w := 0; w < MaxConcurrentWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range proxyChan {
				// Check if context is cancelled
				select {
				case <-ctx.Done():
					log.Println("Scheduler: IP check cancelled - context done")
					return
				default:
					checkSingleProxyIP(p, settings, db, geoIPClient)
				}
			}
		}()
	}

	// Send proxies to workers
	for i := range proxies {
		select {
		case <-ctx.Done():
			log.Println("Scheduler: Stopping IP check - context cancelled")
			close(proxyChan)
			wg.Wait()
			return
		case proxyChan <- &proxies[i]:
		}
	}
	close(proxyChan)

	// Wait for all workers to finish
	wg.Wait()
}

func checkSingleProxyIP(p *Proxy, settings *Settings, db *gorm.DB, geoIPClient *GeoIPClient) {
	log.Printf("Scheduler: Checking IP for proxy %s (%s)", p.Ip, p.Id)

	lastCheck := p.LastCheck

	// 1. Сначала проверяем Ping - если прокси мёртв, нет смысла проверять IP
	latency, err := Ping(settings, p)
	if err != nil {
		log.Printf("Scheduler: Ping failed for proxy %s: %v", p.Ip, err)
		p.Failures++
		p.LastLatency = 0
		if p.Failures > 2 {
			p.LastStatus = 2 // Mark as dead
		}
		// Don't update LastCheck if proxy is dead
	} else {
		// Proxy is alive - update status and check real IP
		p.LastLatency = latency
		p.LastStatus = 1
		p.Failures = 0

		// Calculate uptime only if we have a valid previous check time
		if !lastCheck.IsZero() {
			elapsed := time.Since(lastCheck)
			p.Uptime += int(elapsed.Minutes())
		}
		p.LastCheck = time.Now()

		// Now get real IP (only if proxy is working)
		realIP, realCountry, operator, err := RealIp(settings, p, db, geoIPClient)
		if err != nil {
			log.Printf("Scheduler: Failed to get real IP for proxy %s: %v", p.Ip, err)
			// Continue - don't fail the whole check just because RealIp failed
		} else {
			p.RealIP = realIP
			p.RealCountry = realCountry
			p.Operator = operator
		}
	}

	// Сохраняем обновленный прокси в базе данных.
	if err := p.Save(db); err != nil {
		log.Printf("Scheduler: Error saving updated proxy %s: %v", p.Ip, err)
	}
}

// StartIPCheckScheduler запускает периодическую проверку реальных IP-адресов для всех прокси.
// Он использует WaitGroup для сигнализации о завершении и quit-канал для грациозной остановки.
func StartIPCheckScheduler(wg *sync.WaitGroup, quit <-chan struct{}, db *gorm.DB, settings *Settings, geoIPClient *GeoIPClient, notifier *NotificationService) {
	wg.Add(1)
	defer wg.Done()

	if settings.CheckIPInterval <= 0 {
		log.Println("IP check scheduler is disabled because CheckIPInterval is zero or negative.")
		return
	}

	log.Printf("Starting IP check scheduler. Interval: %d minutes.", settings.CheckIPInterval)

	ticker := time.NewTicker(time.Duration(settings.CheckIPInterval) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Try to acquire lock before starting work
			if !ipCheckMu.TryLock() {
				log.Println("IP check skipped — previous job still running")
				continue
			}

			// Start the check in a goroutine
			go func() {
				defer ipCheckMu.Unlock()

				var proxies []Proxy
				// Загружаем все прокси из базы данных.
				if err := db.Find(&proxies).Error; err != nil {
					log.Println("Scheduler: Error fetching proxies for IP check:", err)
					return
				}

				log.Println("Scheduler: Starting scheduled IP check for all proxies...")
				log.Printf("Scheduler: Found %d proxies to check.", len(proxies))

				// Create context with timeout for the entire check cycle
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(settings.CheckIPInterval)*time.Minute)
				defer cancel()

				// Use notification-enabled iterator
				IPCheckIteratorWithNotifications(ctx, proxies, settings, db, geoIPClient, notifier)

				log.Println("Scheduler: Finished scheduled IP check.")
			}()

		case <-quit:
			log.Println("Scheduler: Shutting down IP check scheduler.")
			return
		}
	}
}

func HealthCheckIterator(proxies []Proxy, settings *Settings, db *gorm.DB) {
	ctx := context.Background()
	HealthCheckIteratorWithContext(ctx, proxies, settings, db)
}

func HealthCheckIteratorWithContext(ctx context.Context, proxies []Proxy, settings *Settings, db *gorm.DB) {
	// Create a worker pool for concurrent checking
	var wg sync.WaitGroup
	proxyChan := make(chan *Proxy, len(proxies))

	// Start worker goroutines
	for w := 0; w < MaxConcurrentWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range proxyChan {
				// Check if context is cancelled
				select {
				case <-ctx.Done():
					log.Println("Scheduler: Health check cancelled - context done")
					return
				default:
					checkSingleProxyHealth(p, settings, db)
				}
			}
		}()
	}

	// Send proxies to workers
	for i := range proxies {
		select {
		case <-ctx.Done():
			log.Println("Scheduler: Stopping health check - context cancelled")
			close(proxyChan)
			wg.Wait()
			return
		case proxyChan <- &proxies[i]:
		}
	}
	close(proxyChan)

	// Wait for all workers to finish
	wg.Wait()
}

func checkSingleProxyHealth(p *Proxy, settings *Settings, db *gorm.DB) {
	log.Printf("Scheduler: Health checking proxy %s (%s)-%s", p.Ip, p.Id, p.Name)

	// Проверяем Speed
	speed, upload, err := CheckSpeed(settings, p, db)
	if err != nil {
		log.Printf("Scheduler: Speed check failed for proxy %s-%s: %v", p.Name, p.Ip, err)
		// Don't update speed on error - keep previous values
	} else {
		// Store speed in Mbps
		p.Speed = int(speed)
		p.Upload = int(upload)
		log.Printf("Scheduler: Speed check completed for proxy %s - Download: %d Mbps, Upload: %d Mbps", p.Ip, p.Speed, p.Upload)
	}

	// Сохраняем обновленные данные
	if err := p.Save(db); err != nil {
		log.Printf("Scheduler: Error saving updated proxy %s: %v", p.Ip, err)
	}
}

// StartHealthCheckScheduler запускает периодическую проверку Ping и Speed для всех прокси.
func StartHealthCheckScheduler(wg *sync.WaitGroup, quit <-chan struct{}, db *gorm.DB, settings *Settings, notifier *NotificationService) {
	wg.Add(1)
	defer wg.Done()

	// Предполагаем, что в настройках есть CheckHealthInterval. Если нет, его нужно добавить.
	// Если интервал не задан, воркер не запускается.
	if settings.SpeedCheckInterval <= 0 {
		log.Println("Health check scheduler is disabled because CheckHealthInterval is zero or negative.")
		return
	}

	log.Printf("Starting Health check scheduler. Interval: %d minutes.", settings.SpeedCheckInterval)

	ticker := time.NewTicker(time.Duration(settings.SpeedCheckInterval) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Try to acquire lock before starting work
			if !healthMu.TryLock() {
				log.Println("Health check skipped — previous job still running")
				continue
			}

			// Start the check in a goroutine
			go func() {
				defer healthMu.Unlock()

				log.Println("Scheduler: Starting scheduled health check for all proxies...")

				var proxies []Proxy
				if err := db.Find(&proxies).Error; err != nil {
					log.Println("Scheduler: Error fetching proxies for health check:", err)
					return
				}

				// Create context with timeout for the entire check cycle
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(settings.SpeedCheckInterval)*time.Minute)
				defer cancel()

				// Use notification-enabled iterator
				HealthCheckIteratorWithNotifications(ctx, proxies, settings, db, notifier)
			}()


		case <-quit:
			log.Println("Scheduler: Shutting down health check scheduler.")
			return
		}
	}
}
