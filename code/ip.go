package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IP struct {
	Ip      string `json:"ip"`
	Country string `json:"country"`
	Cc      string `json:"cc"`
}

func RealIp(stg *Settings, proxy *Proxy, db *gorm.DB, geoIPClient *GeoIPClient) (string, string, string, error) {
	client, err := newProxyClient(proxy, stg)
	if err != nil {
		log.Println("Error creating proxy client:", err)
		return "", "", "", nil
	}

	rsp, err := client.Get("https://api.myip.com")
	if err != nil {
		log.Println(err)
		return "", "", "", nil
	}
	defer rsp.Body.Close()

	ip := &IP{}
	json.NewDecoder(rsp.Body).Decode(ip)

	orerator, err := geoIPClient.ReadData(ip.Ip)
	if err != nil {
		log.Println("Error reading geoIP data:", err)
	}
	op := orerator.ISP
	if strings.Contains(strings.ToLower(op), "moldtelecom") {
		op = "Moldtelecom"
	}

	// get last timestamp 
	var pIpLog ProxyIPLog;
	lastLog, _ := pIpLog.LastByTimestamp(proxy.Id, db)

	stack := false
	if lastLog != nil {
		// Если IP не менялся более 12 часов
		if time.Since(lastLog.Timestamp) > 12*time.Hour && lastLog.Ip == ip.Ip {
			stack = true
		}
	}
	
	fmt.Println("proxy.RealIP: ", ip.Ip, proxy.Ip)
	
	if lastLog != nil && ip.Ip != "" {
		proxy.LastIPChange = time.Now();

		hist := ProxyIPLog{
			Id:         uuid.NewString(),
			ProxyId:    proxy.Id,
			Timestamp:  time.Now(),
			Ip:         ip.Ip,
			OldIp:      lastLog.Ip,
			Country:    ip.Country,
			OldCountry: proxy.RealCountry,
			ISP:        op,
			OldISP:     op,
		}
		if err := hist.Save(db); err != nil {
			log.Println("Error saving IP log:", err)
		}
	} else if ip.Ip != "" && ip.Ip != proxy.Ip {
		proxy.LastIPChange = time.Now()
		hist := ProxyIPLog{
			Id:         uuid.NewString(),
			ProxyId:    proxy.Id,
			Timestamp:  time.Now(),
			Ip:         ip.Ip,
			OldIp:      proxy.Ip,
			Country:    ip.Country,
			OldCountry: proxy.RealCountry,
			ISP:        op,
			OldISP:     op,
			Stack: 			stack,
		}
		if err := hist.Save(db); err != nil {
			log.Println("Error saving IP log:", err)
		}
	} 

	return ip.Ip, ip.Country, orerator.ISP, nil
}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
