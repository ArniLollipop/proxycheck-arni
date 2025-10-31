package main

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

type IP struct {
	Ip      string `json:"ip"`
	Country string `json:"country"`
	Cc      string `json:"cc"`
}

func RealIp(stg *Settings, proxy *Proxy) (string, string) {
	// ip := GetOutboundIP()
	// log.Println(ip)
	// request, err := http.NewRequest("GET", "http://"+ip+"/:8081/test", nil)
	// if err != nil {
	// 	log.Println(err)
	// }

	// trace := &httptrace.ClientTrace{
	// 	GotConn: func(connInfo httptrace.GotConnInfo) {
	// 		fmt.Printf("Got Conn: %+v\n", connInfo)
	// 	},
	// 	DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
	// 		fmt.Printf("DNS Info: %+v\n", dnsInfo)
	// 	},
	// }
	// request = request.WithContext(httptrace.WithClientTrace(request.Context(), trace))

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy: http.ProxyURL(&url.URL{
				Scheme: "http",
				User:   url.UserPassword(proxy.Username, proxy.Password),
				Host:   proxy.Ip + ":" + proxy.Port,
			}),
			// DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// 	conn, err := net.Dial(network, addr)
			// 	log.Println(conn.RemoteAddr().String())
			// 	return conn, err
			// },
		},
		Timeout: time.Duration(stg.Timeout) * time.Second,
		// CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// 	request = req
		// 	return nil
		// },
	}
	// log.Println("http://" + ip + "/:8081/test")
	rsp, err := client.Get("https://api.myip.com")
	if err != nil {
		log.Println(err)
		return "", ""
	}

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
