package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/recoilme/pudge"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var wsMutex sync.Mutex

var (
	dbProxy       *pudge.Db
	dbSettings    *pudge.Db
	dbBench       *BenchData
	client        = Client{}
	settings      *Settings
	benchSettings *BenchSettings
)

func main() {
	var err error
	dbProxy, err = pudge.Open("./database/proxy.db", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer dbProxy.Close()

	dbSettings, err = pudge.Open("./database/settings.db", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer dbSettings.Close()

	dbBench = &BenchData{}
	err = dbBench.Read()
	if err != nil {
		log.Println(err)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("defered exit")
		err := dbBench.Write()
		if err != nil {
			log.Println(err)
		}
		os.Exit(0)
	}()

	settings = SettingsDefault(dbSettings)
	benchSettings = BenchSettingsDefault(dbSettings)

	go func() {
		t := time.Duration(settings.Repeat) * time.Minute
		var diff time.Duration = t

		for {
			time.Sleep(diff)
			now := time.Now()

			keys, err := dbProxy.Keys(nil, 0, 0, true)
			if err != nil {
				log.Println(err)
			}
			for _, key := range keys {
				time.Sleep(200 * time.Millisecond)
				var proxy Proxy
				err := dbProxy.Get(key, &proxy)
				if err != nil {
					log.Println(err)
					continue
				}
				latency, err := Ping(settings, &proxy)
				if err != nil {
					log.Println(err)
				}
				if latency == 0 {
					proxy.Failures += 1
					proxy.LastLatency = 0
					proxy.LastStatus = 2
				} else {
					proxy.LastLatency = latency
					proxy.LastStatus = 1
				}
				proxy.RealIP, proxy.RealCountry = RealIp(settings, &proxy)
				err = proxy.Update(dbProxy)
				if err != nil {
					log.Println(err)
					return
				}
				client.Connect.WriteJSON(&Message{
					Cmd:   "update",
					Value: proxy,
				})
			}
			d := time.Since(now)
			if d < t {
				diff = t - d
			}
		}

	}()

	// go func() {
	// 	http.ListenAndServe(":8081", H{})
	// }()

	log.Println("Runned on http://localhost:8080")
	http.ListenAndServe(":8080", handler{})
}

// type H struct{}

// func (h H) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
// 	if rq.URL.Path == "/" {
// 		id := rq.FormValue("id")
// 		log.Println("id: ", id)
// 		p := &Proxy{}
// 		p.Get(dbProxy, id)
// 		log.Println(*p)
// 		ip, _ := RealIp(p)
// 		log.Println("ip: ", ip)
// 		_, err := rw.Write([]byte(ip))
// 		if err != nil {
// 			log.Println(err)
// 		}
// 		return

// 	} else if rq.URL.Path == "/test" {
// 		log.Println("test path")
// 		log.Println(rq.RemoteAddr)
// 	}
// }
