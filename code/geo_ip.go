package main

import (
	"net"

	"github.com/oschwald/maxminddb-golang"
)

type GeoIPClient struct {
	ispDb *maxminddb.Reader
}

type IpData struct {
	ISP string
}

func (c GeoIPClient) ReadData(ip string) (data IpData, err error) {
	parsdIp := net.ParseIP(ip)
	var isp_record struct {
		Isp               string `maxminddb:"isp"`
		MobileNetworkCode string `maxminddb:"mobile_network_code"`
	}
	if err := c.ispDb.Lookup(parsdIp, &isp_record); err != nil {
		return data, err
	}
	data.ISP = isp_record.Isp

	return
}

func NewGeoIPClient(ispPatch string) (*GeoIPClient, error) {
	idb, err := maxminddb.Open(ispPatch)
	if err != nil {
		return nil, err
	}
	return &GeoIPClient{ispDb: idb}, nil
}
