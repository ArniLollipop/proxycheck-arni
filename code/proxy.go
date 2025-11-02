package main

import (
	"errors"
	"time"
)

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
