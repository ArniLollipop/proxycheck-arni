package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CheckSpeed измеряет скорость загрузки через прокси в КБ/с.
// Для точного измерения скорости рекомендуется использовать URL-адрес
// для `settings.Url`, который отдает файл размером не менее нескольких мегабайт.
func CheckSpeed(settings *Settings, proxy *Proxy, db *gorm.DB) (float64, error) {
	client, err := newProxyClient(proxy, settings)
	if err != nil {
		return 0, err
	}

	startTime := time.Now()
	r, err := client.Get("https://alerts.in.ua")
	if err != nil {
		return 0, err
	}
	duration := time.Since(startTime)
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return 0, errors.New("bad status code: " + r.Status)
	}

	// Читаем тело ответа, чтобы измерить размер и время полной загрузки.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return 0, err
	}
	downloadedBytes := len(body)
	if downloadedBytes == 0 {
		return 0, errors.New("downloaded file is empty")
	}

	if duration.Seconds() == 0 {
		return 0, errors.New("download duration is zero")
	}

	// Рассчитываем скорость в КБ/с.
	// (Количество байт / время в секундах) / 1024
	speedKbps := float64(downloadedBytes*8) / 1024 / duration.Seconds()
	hist := ProxySpeedLog{
		Id:        uuid.NewString(),
		ProxyId:   proxy.Id,
		Timestamp: time.Now(),
		Speed:     int(speedKbps),
	}
	if err := hist.Save(db); err != nil {
		log.Println("Error saving speed log:", err)
	}

	return speedKbps, nil
}

func Ping(settings *Settings, proxy *Proxy) (int, error) {
	client, err := newProxyClient(proxy, settings)
	if err != nil {
		return 0, err
	}

	startTime := time.Now()
	r, err := client.Get(settings.Url)
	if err != nil {
		return 0, err
	}
	defer r.Body.Close()

	if r.StatusCode == 403 || r.StatusCode == 407 {
		return 0, errors.New("status code 403|407")
	}
	diff := time.Since(startTime) / time.Millisecond

	return int(diff), nil
}
