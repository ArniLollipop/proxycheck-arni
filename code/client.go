package main

import (
	"log"
	"math"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Connect *websocket.Conn
}

type Message struct {
	Cmd   string      `json:"cmd"`
	Value interface{} `json:"value"`
}

var isBench bool

func (c *Client) Listen() {
	msg := &Message{}
	for {
		err := c.Connect.ReadJSON(msg)
		if err != nil {
			log.Println(err)
			return
		}
		switch msg.Cmd {
		case "add":
			proxy := &Proxy{}
			proxy.Save(dbProxy, msg.Value.(string))
			wsMutex.Lock()
			client.Connect.WriteJSON(&Message{
				Cmd:   "add",
				Value: proxy,
			})
			wsMutex.Unlock()
			log.Println(*proxy)

		case "start":
			keys, err := dbProxy.Keys(nil, 0, 0, true)
			if err != nil {
				log.Println(err)
			}
			for _, key := range keys {
				var p Proxy
				err := dbProxy.Get(key, &p)
				if err != nil {
					log.Println(err)
					continue
				}
				wsMutex.Lock()
				client.Connect.WriteJSON(&Message{
					Cmd:   "add",
					Value: &p,
				})
				wsMutex.Unlock()
			}

		case "benchInit":
			for _, v := range dbBench.benches {
				wsMutex.Lock()
				client.Connect.WriteJSON(&Message{
					Cmd:   "badd",
					Value: &v,
				})
				wsMutex.Unlock()
			}
		case "benchStart":
			if isBench {
				continue
			}
			isBench = true

			for _, v := range dbBench.benches {
				go func(b Bench) {

					settingsMutex.Lock()
					stg := *benchSettings
					settingsMutex.Unlock()

					autoResetTime := time.Now().Add(time.Duration(stg.Reset) * time.Hour)

					var failStartTime time.Time
					isFail := false

					for {
						time.Sleep(time.Duration(stg.Interval) * time.Millisecond)
						now := time.Now()
						_, err := net.DialTimeout("tcp", b.Addr, time.Duration(stg.Timeout)*time.Millisecond)
						if err != nil {
							// log.Println("start fail: ", b.Addr, err.Error())
						}
						if err != nil && !isFail {
							failStartTime = now
							isFail = true
						}

						if err == nil {
							b.RequestCount++
							diff := time.Since(now)
							b.LatencySum += diff.Milliseconds()
							b.Latency = b.LatencySum / b.RequestCount
							// log.Println(b.ID, b.Latency)
						}
						if err == nil && isFail && isBench {
							isFail = false

							diff := time.Since(failStartTime)
							diffSec := math.Round(diff.Seconds()*10) / 10

							br := BenchResult{
								Start: failStartTime.Format("15:04:05"),
								End:   diffSec,
							}

							b.LastStatus = br.End
							b.LastTime = br.Start
							b.Results = append(b.Results, br)

							b.Results = append(b.Results, br)

							dbBench.mu.Lock()
							dbBench.benches[b.ID] = b
							dbBench.mu.Unlock()

							brsp := BenchResponse{
								ID:           b.ID,
								Addr:         b.Addr,
								RequestCount: b.RequestCount,
								Latency:      b.Latency,
								LastTime:     b.LastTime,
								LastStatus:   b.LastStatus,
							}

							// log.Println(b.ID, b.LastTime, b.LastStatus)

							wsMutex.Lock()
							client.Connect.WriteJSON(&Message{
								Cmd:   "bupdate",
								Value: &brsp,
							})
							wsMutex.Unlock()
						}

						if !isBench {
							log.Println("stop: ", b.Addr)
							return
						}

						if time.Now().After(autoResetTime) {
							b.RequestCount = 0
							b.Latency = 0
							b.LatencySum = 0
							b.LastTime = ""
							b.LastStatus = 0
							b.Results = []BenchResult{}

							dbBench.mu.Lock()
							dbBench.benches[b.ID] = b
							dbBench.mu.Unlock()

							autoResetTime = autoResetTime.Add(time.Duration(stg.Reset) * time.Hour)

						}
					}
				}(v)
			}

		case "benchStop":
			isBench = false
		}
	}
}
