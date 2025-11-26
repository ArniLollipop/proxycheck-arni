package main

import (
	"errors"
	"fmt"
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

	fmt.Println("checkSpeed")

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
	tg.DownloadTest()
	tg.UploadTest()
	upload := tg.ULSpeed.Mbps()
	download := tg.DLSpeed.Mbps()

	downloadKb := download * 1000
	uploadKb := upload * 1000


	proxy.Speed = int(downloadKb);
	proxy.Upload = int(uploadKb);
	
	hist := ProxySpeedLog{
		Id:        uuid.NewString(),
		ProxyId:   proxy.Id,
		Timestamp: time.Now(),
		Speed:     int(downloadKb),
		Upload:    int(uploadKb),
	}
	if err := hist.Save(db); err != nil {
		log.Println("Error saving speed log:", err)
	}

	if err := proxy.Save(db); err != nil {
		log.Println("Error saving proxy's speed log:", err)
	}

	fmt.Println("Saved speed log:", tg.URL)

	return downloadKb, uploadKb, nil
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
