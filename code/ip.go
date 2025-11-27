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
		log.Printf("Error creating proxy client for %s:%s - %v", proxy.Ip, proxy.Port, err)
		return "", "", "", err
	}

	rsp, err := client.Get("https://api.myip.com")
	if err != nil {
		log.Printf("Error getting real IP for %s:%s - %v", proxy.Ip, proxy.Port, err)
		return "", "", "", err
	}
	defer rsp.Body.Close()

	ip := &IP{}
	if err := json.NewDecoder(rsp.Body).Decode(ip); err != nil {
		log.Printf("Error decoding IP response for %s:%s - %v", proxy.Ip, proxy.Port, err)
		return "", "", "", err
	}

	operator, err := geoIPClient.ReadData(ip.Ip)
	if err != nil {
		log.Printf("Error reading geoIP data for %s - %v", ip.Ip, err)
		// Continue with empty operator instead of failing
	}
	op := operator.ISP
	if strings.Contains(strings.ToLower(op), "moldtelecom") {
		op = "Moldtelecom"
	}

	// Get last IP log entry
	var pIpLog ProxyIPLog
	lastLog, err := pIpLog.LastByTimestamp(proxy.Id, db)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Error fetching last IP log for proxy %s - %v", proxy.Id, err)
	}

	// Detect if IP is stuck (not changed for more than 12 hours)
	stack := false
	if lastLog != nil && lastLog.Ip == ip.Ip && time.Since(lastLog.Timestamp) > 12*time.Hour {
		stack = true
		log.Printf("Warning: IP stuck for proxy %s:%s - Same IP %s for >12 hours", proxy.Ip, proxy.Port, ip.Ip)
	}

	// Update proxy Stack field
	proxy.Stack = stack

	// Save IP log only if IP actually changed
	if lastLog != nil && ip.Ip != "" && lastLog.Ip != ip.Ip {
		proxy.LastIPChange = time.Now()
		proxy.Stack = false // IP changed, so not stuck anymore

		hist := ProxyIPLog{
			Id:         uuid.NewString(),
			ProxyId:    proxy.Id,
			Timestamp:  time.Now(),
			Ip:         ip.Ip,
			OldIp:      lastLog.Ip,
			Country:    ip.Country,
			OldCountry: lastLog.Country,
			ISP:        op,
			OldISP:     lastLog.ISP,
			Stack:      false, // IP changed, so not stuck
		}
		if err := hist.Save(db); err != nil {
			log.Printf("Error saving IP log for proxy %s - %v", proxy.Id, err)
		} else {
			log.Printf("IP changed for proxy %s:%s: %s -> %s", proxy.Ip, proxy.Port, lastLog.Ip, ip.Ip)
		}
	} else if lastLog == nil && ip.Ip != "" {
		// First time checking this proxy - create initial log
		proxy.LastIPChange = time.Now()
		proxy.Stack = false
		hist := ProxyIPLog{
			Id:         uuid.NewString(),
			ProxyId:    proxy.Id,
			Timestamp:  time.Now(),
			Ip:         ip.Ip,
			OldIp:      proxy.Ip,
			Country:    ip.Country,
			OldCountry: "",
			ISP:        op,
			OldISP:     "",
			Stack:      false,
		}
		if err := hist.Save(db); err != nil {
			log.Printf("Error saving initial IP log for proxy %s - %v", proxy.Id, err)
		}
	}

	return ip.Ip, ip.Country, operator.ISP, nil
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
