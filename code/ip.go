package main

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IP struct {
	Ip      string `json:"ip"`
	Country string `json:"country"`
	Cc      string `json:"cc"`
}

func RealIp(stg *Settings, proxy *Proxy, db *gorm.DB, geoIPClient *GeoIPClient) (string, string, string) {
	client, err := newProxyClient(proxy, stg)
	if err != nil {
		log.Println("Error creating proxy client:", err)
		return "", "", ""
	}

	rsp, err := client.Get("https://api.myip.com")
	if err != nil {
		log.Println(err)
		return "", "", ""
	}
	defer rsp.Body.Close()

	ip := &IP{}

	orerator, err := geoIPClient.ReadData(ip.Ip)
	if err != nil {
		log.Println("Error reading geoIP data:", err)
	}

	json.NewDecoder(rsp.Body).Decode(ip)
	hist := ProxyIPLog{
		Id:         uuid.NewString(),
		ProxyId:    proxy.Id,
		Timestamp:  time.Now(),
		Ip:         ip.Ip,
		OldIp:      proxy.Ip,
		Country:    ip.Country,
		OldCountry: proxy.RealCountry,
		ISP:        orerator.ISP,
		OldISP:     proxy.Operator,
	}
	if err := hist.Save(db); err != nil {
		log.Println("Error saving IP log:", err)
	}

	return ip.Ip, ip.Country, orerator.ISP
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
