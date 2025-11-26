package main

import (
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// StartIPCheckScheduler запускает периодическую проверку реальных IP-адресов для всех прокси.
// Он использует WaitGroup для сигнализации о завершении и quit-канал для грациозной остановки.
func StartIPCheckScheduler(wg *sync.WaitGroup, quit <-chan struct{}, db *gorm.DB, settings *Settings, geoIPClient *GeoIPClient) {
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
			log.Println("Scheduler: Starting scheduled IP check for all proxies...")

			var proxies []Proxy
			// Загружаем все прокси из базы данных.
			if err := db.Find(&proxies).Error; err != nil {
				log.Println("Scheduler: Error fetching proxies for IP check:", err)
				continue // Пропускаем эту итерацию, ждем следующего тика.
			}

			log.Printf("Scheduler: Found %d proxies to check.", len(proxies))

			// Проходим по каждому прокси.
			for _, p := range proxies {
				// Важно: используем &proxies[i], чтобы работать с оригинальным элементом слайса, а не с его копией.

				log.Printf("Scheduler: Checking IP for proxy %s (%s)", p.Ip, p.Id)

				// Вызываем существующую функцию для получения реального IP.
				realIP, realCountry, operator := RealIp(settings, &p, db, geoIPClient)
				p.LastCheck = time.Now()

				// Обновляем поля в объекте прокси.
				p.RealIP = realIP
				p.RealCountry = realCountry
				p.Operator = operator

				// 1. Проверяем Ping
				latency, err := Ping(settings, &p)
				if err != nil || p.LastStatus == 2 {
					log.Printf("Scheduler: Ping failed for proxy %s: %v", p.Ip, err)
					p.Failures++;
					p.Uptime = 0;
					p.LastLatency = 0;
					if p.Failures > 2 {
						p.LastStatus = 2
					}
				} else {
					p.LastLatency = latency
					p.LastStatus = 1
					p.Failures = 0
					
					if p.LastCheck.IsZero() {
						p.LastCheck = time.Now().Add(-10 * time.Minute)
					}
					uptime := time.Since(p.LastCheck).Minutes()
					p.Uptime += int(uptime)
				}

				// Сохраняем обновленный прокси в базе данных.
				if err := p.Save(db); err != nil {
					log.Printf("Scheduler: Error saving updated proxy %s: %v", p.Ip, err)
				}
			}
			log.Println("Scheduler: Finished scheduled IP check.")

		case <-quit:
			log.Println("Scheduler: Shutting down IP check scheduler.")
			return
		}
	}
}

// StartHealthCheckScheduler запускает периодическую проверку Ping и Speed для всех прокси.
func StartHealthCheckScheduler(wg *sync.WaitGroup, quit <-chan struct{}, db *gorm.DB, settings *Settings) {
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
			log.Println("Scheduler: Starting scheduled health check for all proxies...")

			var proxies []Proxy
			if err := db.Find(&proxies).Error; err != nil {
				log.Println("Scheduler: Error fetching proxies for health check:", err)
				continue
			}

			log.Printf("Scheduler: Found %d proxies for health check.", len(proxies))

			for _, p := range proxies {
				log.Printf("Scheduler: Health checking proxy %s (%s)", p.Ip, p.Id)

				// 2. Проверяем Speed
				speed, upload, err := CheckSpeed(settings, &p, db)
				if err != nil {
					log.Printf("Scheduler: Speed check failed for proxy %s: %v", p.Ip, err)
				} else {
					p.Speed = int(speed)
					if p.Speed == 0 {
						p.Speed = 1
					}
					p.Upload = int(upload)
					if p.Upload == 0 {
						p.Upload = 1
					}
				}

				// Сохраняем обновленные данные
				if err := p.Save(db); err != nil {
					log.Printf("Scheduler: Error saving updated proxy %s: %v", p.Ip, err)
				}
			}
			log.Println("Scheduler: Finished scheduled health check.")

		case <-quit:
			log.Println("Scheduler: Shutting down health check scheduler.")
			return
		}
	}
}
