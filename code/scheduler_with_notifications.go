package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IPCheckIteratorWithNotifications checks proxies with notification support
func IPCheckIteratorWithNotifications(ctx context.Context, proxies []Proxy, settings *Settings, db *gorm.DB, geoIPClient *GeoIPClient, notifier *NotificationService) {
	var wg sync.WaitGroup
	proxyChan := make(chan *Proxy, len(proxies))

	// Start worker goroutines
	for w := 0; w < MaxConcurrentWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range proxyChan {
				select {
				case <-ctx.Done():
					log.Println("Scheduler: IP check cancelled - context done")
					return
				default:
					checkSingleProxyIPWithNotifications(p, settings, db, geoIPClient, notifier)
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

	wg.Wait()
}

// checkSingleProxyIPWithNotifications checks a single proxy with notifications
func checkSingleProxyIPWithNotifications(p *Proxy, settings *Settings, db *gorm.DB, geoIPClient *GeoIPClient, notifier *NotificationService) {
	log.Printf("Scheduler: Checking IP for proxy %s (%s)", p.Ip, p.Id)

	lastCheck := p.LastCheck
	wasDown := p.LastStatus == 2 // Remember if proxy was down before

	// 1. Check Ping first
	latency, err := Ping(settings, p)
	if err != nil {
		log.Printf("Scheduler: Ping failed for proxy %s: %v", p.Ip, err)
		p.Failures++
		p.LastLatency = 0

		// Log failure
		failureLog := &ProxyFailureLog{
			ID:        uuid.NewString(),
			ProxyID:   p.Id,
			Timestamp: time.Now(),
			ErrorType: "ping_failed",
			ErrorMsg:  err.Error(),
			Latency:   p.LastLatency,
		}
		if err := failureLog.Save(db); err != nil {
			log.Printf("Failed to save failure log: %v", err)
		}

		if p.Failures > 2 {
			p.LastStatus = 2 // Mark as dead

			// Notify if proxy went down
			if !wasDown && settings.NotifyOnDown {
				notifier.NotifyProxyDown(p, err.Error())
			}
		}
	} else {
		// Proxy is alive
		p.LastLatency = latency
		p.LastStatus = 1
		p.Failures = 0

		// Calculate uptime
		if !lastCheck.IsZero() {
			elapsed := time.Since(lastCheck)
			p.Uptime += int(elapsed.Minutes())
		}
		p.LastCheck = time.Now()

		// Notify if proxy recovered
		if wasDown && settings.NotifyOnRecovery {
			notifier.NotifyProxyRecovered(p)
		}

		// Now get real IP (only if proxy is working)
		realIP, realCountry, operator, err := RealIp(settings, p, db, geoIPClient)
		if err != nil {
			log.Printf("Scheduler: Failed to get real IP for proxy %s: %v", p.Ip, err)

			// Log IP check failure
			failureLog := &ProxyFailureLog{
				ID:        uuid.NewString(),
				ProxyID:   p.Id,
				Timestamp: time.Now(),
				ErrorType: "ip_check_failed",
				ErrorMsg:  err.Error(),
				Latency:   latency,
			}
			if err := failureLog.Save(db); err != nil {
				log.Printf("Failed to save failure log: %v", err)
			}
		} else {
			// Check if IP changed
			oldIP := p.RealIP
			if oldIP != "" && oldIP != realIP && settings.NotifyOnIPChange {
				notifier.NotifyIPChanged(p, oldIP, realIP)
			}

			p.RealIP = realIP
			p.RealCountry = realCountry
			p.Operator = operator

			// Check if IP is stuck (>24 hours)
			if p.Stack && settings.NotifyOnIPStuck {
				// Calculate hours stuck
				var lastLog ProxyIPLog
				err := db.Where("proxy_id = ?", p.Id).
					Order("timestamp desc").
					Limit(1).
					First(&lastLog).Error

				if err == nil {
					hours := int(time.Since(lastLog.Timestamp).Hours())
					if hours >= 24 {
						notifier.NotifyIPStuck(p, realIP, hours)
					}
				}
			}
		}
	}

	// Save updated proxy
	if err := p.Save(db); err != nil {
		log.Printf("Scheduler: Error saving updated proxy %s: %v", p.Ip, err)
	}
}

// HealthCheckIteratorWithNotifications checks proxy speeds with notifications
func HealthCheckIteratorWithNotifications(ctx context.Context, proxies []Proxy, settings *Settings, db *gorm.DB, notifier *NotificationService) {
	var wg sync.WaitGroup
	proxyChan := make(chan *Proxy, len(proxies))

	// Start worker goroutines
	for w := 0; w < MaxConcurrentWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range proxyChan {
				select {
				case <-ctx.Done():
					log.Println("Scheduler: Health check cancelled - context done")
					return
				default:
					checkSingleProxyHealthWithNotifications(p, settings, db, notifier)
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

	wg.Wait()
}

// checkSingleProxyHealthWithNotifications checks speed with notifications
func checkSingleProxyHealthWithNotifications(p *Proxy, settings *Settings, db *gorm.DB, notifier *NotificationService) {
	log.Printf("Scheduler: Health checking proxy %s (%s)-%s", p.Ip, p.Id, p.Name)

	speed, upload, err := CheckSpeed(settings, p, db)
	if err != nil {
		log.Printf("Scheduler: Speed check failed for proxy %s-%s: %v", p.Name, p.Ip, err)

		// Log speed check failure
		failureLog := &ProxyFailureLog{
			ID:        uuid.NewString(),
			ProxyID:   p.Id,
			Timestamp: time.Now(),
			ErrorType: "speed_check_failed",
			ErrorMsg:  err.Error(),
			Latency:   p.LastLatency,
		}
		if err := failureLog.Save(db); err != nil {
			log.Printf("Failed to save failure log: %v", err)
		}
	} else {
		// Store speed in Mbps
		p.Speed = int(speed)
		p.Upload = int(upload)

		// Check if speed is below threshold
		if settings.NotifyOnLowSpeed && p.Speed < settings.LowSpeedThreshold && p.Speed > 0 {
			notifier.NotifyLowSpeed(p, settings.LowSpeedThreshold)
		}

		log.Printf("Scheduler: Speed check completed for proxy %s - Download: %d Mbps, Upload: %d Mbps", p.Ip, p.Speed, p.Upload)
	}

	// Save updated data
	if err := p.Save(db); err != nil {
		log.Printf("Scheduler: Error saving updated proxy %s: %v", p.Ip, err)
	}
}
