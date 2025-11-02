package main

import (
	"encoding/json"
	"log"
	"net"
)

type IP struct {
	Ip      string `json:"ip"`
	Country string `json:"country"`
	Cc      string `json:"cc"`
}

func RealIp(stg *Settings, proxy *Proxy) (string, string) {
	client, err := newProxyClient(proxy, stg)
	if err != nil {
		log.Println("Error creating proxy client:", err)
		return "", ""
	}

	rsp, err := client.Get("https://api.myip.com")
	if err != nil {
		log.Println(err)
		return "", ""
	}
	defer rsp.Body.Close()

	ip := &IP{}

	json.NewDecoder(rsp.Body).Decode(ip)

	log.Println(*ip)

	return ip.Ip, ip.Country
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
