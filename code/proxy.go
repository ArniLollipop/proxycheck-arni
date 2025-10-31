package main

import (
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/recoilme/pudge"
)

type Proxy struct {
	Id          int    `json:"id"`
	Ip          string `json:"ip"`
	Port        string `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	LastLatency int    `json:"lastLatency"`
	Tag         string `json:"tag"`
	LastStatus  int    `json:"lastStatus"`
	Failures    int    `json:"failures"`
	RealIP      string `json:"realIP"`
	RealCountry string `json:"realCountry"`
}

func (p *Proxy) Parse(proxy string) {
	parts := strings.Split(proxy, ":")
	p.Ip = parts[0]
	p.Port = parts[1]
	p.Username = parts[2]
	p.Password = parts[3]

	if len(parts) == 5 {
		isIP := strings.Split(parts[4], ".")
		if len(isIP) == 4 {
			p.RealIP = parts[5]
		} else {
			p.Tag = parts[4]
		}
	}
	if len(parts) == 6 {
		p.RealIP = parts[5]
	}
}

func (p *Proxy) Save(db *pudge.Db, proxy string) error {
	var stg Settings
	dbSettings.Get("settings", &stg)
	stg.LastIndex++
	dbSettings.Set("settings", &stg)
	p.Parse(proxy)
	p.Id = stg.LastIndex
	p.RealIP, p.RealCountry = RealIp(settings, p)
	err := db.Set(p.Id, p)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (p *Proxy) Update(db *pudge.Db) error {
	err := db.Set(p.Id, p)
	if err != nil {
		return err
	}
	return nil
}

func (p *Proxy) String() string {
	s := p.Ip + ":" + p.Port + ":" + p.Username + ":" + p.Password
	if p.Tag != "" {
		s += ":" + p.Tag
	}
	if p.RealIP != "" {
		s += ":" + p.RealIP
	}
	return s
}

func (p *Proxy) Get(db *pudge.Db, id string) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
		return
	}
	err = db.Get(idInt, p)
	if err != nil {
		return
	}
}

func Ping(settings *Settings, proxy *Proxy) (int, error) {
	settingsMutex.RLock()
	stg := *settings
	settingsMutex.RUnlock()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy: http.ProxyURL(&url.URL{
				Scheme: "htts",
				User:   url.UserPassword(proxy.Username, proxy.Password),
				Host:   proxy.Ip + ":" + proxy.Port,
			}),
		},
		Timeout: time.Duration(stg.Timeout) * time.Second,
	}

	startTime := time.Now()
	r, err := client.Get(stg.Url)
	if err != nil {
		return 0, err
	}

	// defer r.Body.Close()
	// b, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	log.Println(err)
	// }
	// log.Println(string(b))

	if r.StatusCode == 403 || r.StatusCode == 407 {
		return 0, errors.New("status code 403|407")
	}

	diff := time.Since(startTime) / time.Millisecond

	return int(diff), nil
}
