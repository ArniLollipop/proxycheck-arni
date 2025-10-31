package main

import (
	"sync"

	"github.com/recoilme/pudge"
)

type Settings struct {
	Url       string `json:"url"`
	Timeout   int    `json:"timeout"`
	Threads   int    `json:"threads"`
	Repeat    int    `json:"repeat"`
	LastIndex int
}

var settingsMutex sync.RWMutex

func SettingsDefault(db *pudge.Db) *Settings {
	stg := &Settings{
		Url:     "https://example.com",
		Timeout: 5,
		Threads: 10,
		Repeat:  15,
	}
	db.Get("settings", stg)
	return stg
}

type BenchSettings struct {
	Timeout  int `json:"timeout"`
	Interval int `json:"interval"`
	Reset    int `json:"reset"`
}

func BenchSettingsDefault(db *pudge.Db) *BenchSettings {
	stg := &BenchSettings{
		Timeout:  500,
		Interval: 200,
		Reset:    24,
	}
	db.Get("benchSettings", stg)
	return stg
}
