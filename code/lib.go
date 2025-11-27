package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// newProxyClient создает и настраивает http.Client для работы через прокси.
func newProxyClient(proxy *Proxy, stg *Settings) (*http.Client, error) {
	// Формируем URL прокси с данными для аутентификации, если они есть.
	proxyStr := fmt.Sprintf("http://%s:%s", proxy.Ip, proxy.Port)

	if proxy.Username != "" {
		proxyStr = fmt.Sprintf("http://%s:%s@%s:%s", proxy.Username, proxy.Password, proxy.Ip, proxy.Port)
	}

	proxyUrl, err := url.Parse(proxyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxy URL: %w", err)
	}

	// Создаем транспорт с настройками прокси.
	transport := &http.Transport{
		Proxy:           http.ProxyURL(proxyUrl),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: stg.SkipSSLVerify},
	}

	// Создаем HTTP-клиент.
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(stg.Timeout) * time.Second,
	}

	return client, nil
}

func sliceStrToIntConvert(slice []string) []int {
	var sliceNew []int
	for _, v := range slice {
		n, err := strconv.Atoi(v)
		if err != nil {
			log.Println(err)
			continue
		}
		sliceNew = append(sliceNew, n)
	}
	return sliceNew
}
