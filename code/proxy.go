package main

import (
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/showwin/speedtest-go/speedtest"
	"gorm.io/gorm"
)

// CheckSpeed измеряет скорость загрузки через прокси в КБ/с.
// Для точного измерения скорости рекомендуется использовать URL-адрес
// для `settings.Url`, который отдает файл размером не менее нескольких мегабайт.
func CheckSpeed(settings *Settings, proxy *Proxy, db *gorm.DB) (float64, float64, error) {
	client, err := newProxyClient(proxy, settings)
	if err != nil {
		return 0, 0, err
	}
	var speedtestClient = speedtest.New(speedtest.WithDoer(client))
	serverList, _ := speedtestClient.FetchServers()
	targets, _ := serverList.FindServer([]int{})
	if len(targets) == 0 {
		return 0, 0, errors.New("no suitable servers found")
	}
	tg := targets[0]

	// Run ping test with callback
	err = tg.PingTest(func(latency time.Duration) {})
	if err != nil {
		return 0, 0, err
	}

	tg.DownloadTest()
	tg.UploadTest()

	ping := float64(tg.Latency.Milliseconds())
	upload := tg.ULSpeed.Mbps()
	download := tg.DLSpeed.Mbps()

	// Retry once if download or upload is 0
	if download == 0 || upload == 0 {
		log.Printf("Speedtest retry for %s:%s - Download: %.2f Mbps, Upload: %.2f Mbps (retrying once)",
			proxy.Ip, proxy.Port, download, upload)

		tg.DownloadTest()
		tg.UploadTest()

		upload = tg.ULSpeed.Mbps()
		download = tg.DLSpeed.Mbps()
	}

	log.Printf("Speedtest results for %s:%s - Ping: %.2fms, Download: %.2f Mbps, Upload: %.2f Mbps",
		proxy.Ip, proxy.Port, ping, download, upload)

	// Store speed in Mbps (not Kbps) and ping in ms
	proxy.Speed = int(download)
	proxy.Upload = int(upload)
	proxy.LastLatency = int(ping)

	hist := ProxySpeedLog{
		Id:        uuid.NewString(),
		ProxyId:   proxy.Id,
		Timestamp: time.Now(),
		Ping:			 ping,
		Speed:     int(download),
		Upload:    int(upload),
	}
	if err := hist.Save(db); err != nil {
		log.Printf("Error saving speed log for proxy %s:%s - %v", proxy.Ip, proxy.Port, err)
	}

	if err := proxy.Save(db); err != nil {
		log.Printf("Error saving proxy speed for %s:%s - %v", proxy.Ip, proxy.Port, err)
	}

	return download, upload, nil
}

func Ping(settings *Settings, proxy *Proxy) (int, error) {
	client, err := newProxyClient(proxy, settings)
	if err != nil {
		return 0, err
	}

	var speedtestClient = speedtest.New(speedtest.WithDoer(client))
	serverList, _ := speedtestClient.FetchServers()
	targets, _ := serverList.FindServer([]int{})
	if len(targets) == 0 {
		return 0, errors.New("no suitable servers found")
	}
	tg := targets[0]

	// Run ping test with callback
	err = tg.PingTest(func(latency time.Duration) {})
	if err != nil {
		return 0, err
	}

	ping := int(tg.Latency.Milliseconds())

	return ping, nil
}
