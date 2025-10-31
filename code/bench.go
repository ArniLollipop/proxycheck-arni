package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

type BenchResponse struct {
	ID           int64   `json:"id"`
	Addr         string  `json:"addr"`
	RequestCount int64   `json:"requestCount"`
	Latency      int64   `json:"latency"`
	LastTime     string  `json:"lastTime"`
	LastStatus   float64 `json:"lastStatus"`
}

type Bench struct {
	ID           int64         `json:"id"`
	Addr         string        `json:"addr"`
	RequestCount int64         `json:"requestCount"`
	LatencySum   int64         `json:"latencySum"`
	Latency      int64         `json:"latency"`
	LastTime     string        `json:"lastTime"`
	LastStatus   float64       `json:"lastStatus"`
	Results      []BenchResult `json:"results"`
}

type BenchResult struct {
	Start string  `json:"start"`
	End   float64 `json:"end"`
}

type BenchData struct {
	benches map[int64]Bench `json:"benches"`
	mu      sync.Mutex      `json:"-"`
}

func (bd *BenchData) Read() error {
	f, err := os.OpenFile("./database/bench.db", os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	bd.benches = make(map[int64]Bench)

	if len(data) > 0 {
		err = json.Unmarshal(data, &bd.benches)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (bd *BenchData) Write() error {
	data, err := json.Marshal(&bd.benches)
	if err != nil {
		log.Println(err)
		return err
	}

	err = ioutil.WriteFile("./database/bench.db", data, 0755)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
