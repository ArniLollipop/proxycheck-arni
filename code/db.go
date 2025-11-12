package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Proxy struct {
	Id           string    `json:"id"`
	Ip           string    `json:"ip"`
	Port         string    `json:"port"`
	Username     string    `json:"username"`
	Password     string    `json:"password"`
	LastLatency  int       `json:"lastLatency"`
	Tag          string    `json:"tag"`
	LastStatus   int       `json:"lastStatus"`
	Failures     int       `json:"failures"`
	RealIP       string    `json:"realIP"`
	RealCountry  string    `json:"realCountry"`
	Contacts     string    `json:"contacts"`
	LastIPChange time.Time `json:"last_ip_change"`
	Operator     string    `json:"operator"`
	Phone        string    `json:"phone"`
	Speed        int       `json:"speed"`
	Upload       int       `json:"upload"`
	Name         string    `json:"name"`
	Uptime       int       `json:"uptime"`
	LastCheck    time.Time `json:"last_check"`
}

func (s *Proxy) Save(db *gorm.DB) error {
	err := db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&s).Error
	return err
}
func (s *Proxy) List(db *gorm.DB) ([]Proxy, error) {
	proxy := []Proxy{}
	err := db.Model(s).Find(&proxy).Error
	return proxy, err
}

func (s *Proxy) Delete(db *gorm.DB) error {
	err := db.Delete(&s).Error
	return err
}

func (s *Proxy) Get(db *gorm.DB, id string) error {
	return db.Where("id =?", id).First(&s).Error
}

func (p *Proxy) Parse(proxy string) {
	proxy = strings.TrimSpace(proxy)
	proxy = strings.TrimPrefix(proxy, "http://")
	proxy = strings.TrimPrefix(proxy, "https://")

	if strings.Contains(proxy, "@") {
		// Format: username:password@ip:port
		parts := strings.SplitN(proxy, "@", 2)
		userpass := strings.SplitN(parts[0], ":", 2)
		hostport := strings.SplitN(parts[1], ":", 2)

		if len(userpass) == 2 {
			p.Username = userpass[0]
			p.Password = userpass[1]
		}
		if len(hostport) == 2 {
			p.Ip = hostport[0]
			p.Port = hostport[1]
		}
	} else {
		// Format: ip:port or ip:port:username:password or username:password:ip:port
		parts := strings.Split(proxy, ":")
		if len(parts) == 4 {
			// Assuming username:password:ip:port based on your examples
			p.Username = parts[0]
			p.Password = parts[1]
			p.Ip = parts[2]
			p.Port = parts[3]
		} else if len(parts) == 2 {
			// Format: ip:port
			p.Ip = parts[0]
			p.Port = parts[1]
		}
	}

	if p.Ip != "" && net.ParseIP(p.Ip) == nil {
		// p.Ip is not a valid IP, so it's probably a domain name. Let's resolve it.
		ips, err := net.LookupIP(p.Ip)
		if err != nil {
			log.Printf("Could not resolve host: %s, error: %v", p.Ip, err)
		} else if len(ips) > 0 {
			// Take the first resolved IP
			p.Ip = ips[0].String()
		}
	}
}

func (s *Proxy) String() string {
	if s.Username != "" && s.Password != "" {
		return fmt.Sprintf("%s:%s@%s:%s", s.Username, s.Password, s.Ip, s.Port)
	}
	return fmt.Sprintf("%s:%s", s.Ip, s.Port)
}

type ProxyVisitLogs struct {
	Id        string    `json:"id"`
	ProxyId   string    `json:"proxy_id"`
	Timestamp time.Time `json:"timestamp"`
	SourceIP  string    `json:"source_ip"`
	TargetIP  string    `json:"target_ip"`
	Domain    string    `json:"domain"`
}

func (s *ProxyVisitLogs) Save(db *gorm.DB) error {
	err := db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&s).Error
	return err
}

type ProxyVisitLogsFilters struct {
	ProxyId   string
	SourceIP  string
	TargetIP  string
	Domain    string
	StartDate time.Time
	EndDate   time.Time
	Page      int
	PageSize  int
	SortField string
}

type scopeFun func(db *gorm.DB) *gorm.DB

func (p *ProxyVisitLogs) buildWhereProxyId(proxyId string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("proxy_id = ?", proxyId)
	}
}

func (p *ProxyVisitLogs) buildWhereSourceIP(sourceIP string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("source_ip LIKE ?", "%"+sourceIP+"%")
	}
}

func (p *ProxyVisitLogs) buildWhereTargetIP(targetIP string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("target_ip LIKE ?", "%"+targetIP+"%")
	}
}

func (p *ProxyVisitLogs) buildWhereDomain(domain string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("domain LIKE ?", "%"+domain+"%")
	}
}

func (p *ProxyVisitLogs) buildWhereBetweenDates(startDate, endDate time.Time) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("timestamp BETWEEN ? AND ?", startDate, endDate)
	}
}

func (p *ProxyVisitLogs) buildWhereMoreThenStartDate(startDate time.Time) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("timestamp > ?", startDate)
	}
}

func (p *ProxyVisitLogs) buildWhereLessThenEndDate(endDate time.Time) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("timestamp < ?", endDate)
	}
}

func (p *ProxyVisitLogs) buildOrder(condition string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(condition)
	}
}

func (p *ProxyVisitLogs) buildConditions(filters ProxyVisitLogsFilters) []func(*gorm.DB) *gorm.DB {
	scopes := make([]func(*gorm.DB) *gorm.DB, 0)

	// ProxyId filter
	if filters.ProxyId != "" {
		scopes = append(scopes, p.buildWhereProxyId(filters.ProxyId))
	}

	// Date filter
	if !filters.StartDate.IsZero() || !filters.EndDate.IsZero() {
		switch {
		case !filters.StartDate.IsZero() && !filters.EndDate.IsZero():
			scopes = append(scopes, p.buildWhereBetweenDates(filters.StartDate, filters.EndDate))
		case !filters.StartDate.IsZero():
			scopes = append(scopes, p.buildWhereMoreThenStartDate(filters.StartDate))
		case !filters.EndDate.IsZero():
			scopes = append(scopes, p.buildWhereLessThenEndDate(filters.EndDate))
		}
	}

	// SourceIP filter
	if filters.SourceIP != "" {
		scopes = append(scopes, p.buildWhereSourceIP(filters.SourceIP))
	}

	// TargetIP filter
	if filters.TargetIP != "" {
		scopes = append(scopes, p.buildWhereTargetIP(filters.TargetIP))
	}

	// Domain filter
	if filters.Domain != "" {
		scopes = append(scopes, p.buildWhereDomain(filters.Domain))
	}

	// Sort
	if filters.SortField != "" {
		scopes = append(scopes, p.buildOrder(filters.SortField))
	} else {
		scopes = append(scopes, p.buildOrder("timestamp desc"))
	}

	return scopes
}

func (p *ProxyVisitLogs) buildConditionsCount(filters ProxyVisitLogsFilters) []func(*gorm.DB) *gorm.DB {
	scopes := make([]func(*gorm.DB) *gorm.DB, 0)

	// ProxyId filter
	if filters.ProxyId != "" {
		scopes = append(scopes, p.buildWhereProxyId(filters.ProxyId))
	}

	// Date filter
	if !filters.StartDate.IsZero() || !filters.EndDate.IsZero() {
		switch {
		case !filters.StartDate.IsZero() && !filters.EndDate.IsZero():
			scopes = append(scopes, p.buildWhereBetweenDates(filters.StartDate, filters.EndDate))
		case !filters.StartDate.IsZero():
			scopes = append(scopes, p.buildWhereMoreThenStartDate(filters.StartDate))
		case !filters.EndDate.IsZero():
			scopes = append(scopes, p.buildWhereLessThenEndDate(filters.EndDate))
		}
	}

	// SourceIP filter
	if filters.SourceIP != "" {
		scopes = append(scopes, p.buildWhereSourceIP(filters.SourceIP))
	}

	// TargetIP filter
	if filters.TargetIP != "" {
		scopes = append(scopes, p.buildWhereTargetIP(filters.TargetIP))
	}

	// Domain filter
	if filters.Domain != "" {
		scopes = append(scopes, p.buildWhereDomain(filters.Domain))
	}

	return scopes
}

func (p *ProxyVisitLogs) List(filters ProxyVisitLogsFilters, db *gorm.DB) ([]ProxyVisitLogs, int64, error) {
	logs := []ProxyVisitLogs{}
	limit := filters.PageSize
	offset := 0

	if filters.Page > 1 {
		offset = limit * (filters.Page - 1)
	}

	scopes := p.buildConditions(filters)
	err := db.Model(p).Scopes(scopes...).Limit(limit).Offset(offset).Find(&logs).Error
	var count int64
	err = db.Model(p).Scopes(scopes...).Count(&count).Error
	return logs, count, err
}

type ProxySpeedLog struct {
	Id        string    `json:"id"`
	ProxyId   string    `json:"proxy_id"`
	Timestamp time.Time `json:"timestamp"`
	Speed     int       `json:"speed"`
	Upload    int       `json:"upload"`
}

func (s *ProxySpeedLog) Save(db *gorm.DB) error {
	err := db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&s).Error
	return err
}

type ProxySpeedLogFilters struct {
	ProxyId   string
	StartDate time.Time
	EndDate   time.Time
	Page      int
	PageSize  int
	SortField string
}

func (p *ProxySpeedLog) buildWhereProxyId(proxyId string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("proxy_id = ?", proxyId)
	}
}

func (p *ProxySpeedLog) buildWhereBetweenDates(startDate, endDate time.Time) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("timestamp BETWEEN ? AND ?", startDate, endDate)
	}
}

func (p *ProxySpeedLog) buildWhereMoreThenStartDate(startDate time.Time) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("timestamp > ?", startDate)
	}
}

func (p *ProxySpeedLog) buildWhereLessThenEndDate(endDate time.Time) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("timestamp < ?", endDate)
	}
}

func (p *ProxySpeedLog) buildOrder(condition string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(condition)
	}
}

func (p *ProxySpeedLog) buildConditions(filters ProxySpeedLogFilters) []func(*gorm.DB) *gorm.DB {
	scopes := make([]func(*gorm.DB) *gorm.DB, 0)

	// ProxyId filter
	if filters.ProxyId != "" {
		scopes = append(scopes, p.buildWhereProxyId(filters.ProxyId))
	}

	// Date filter
	if !filters.StartDate.IsZero() || !filters.EndDate.IsZero() {
		switch {
		case !filters.StartDate.IsZero() && !filters.EndDate.IsZero():
			scopes = append(scopes, p.buildWhereBetweenDates(filters.StartDate, filters.EndDate))
		case !filters.StartDate.IsZero():
			scopes = append(scopes, p.buildWhereMoreThenStartDate(filters.StartDate))
		case !filters.EndDate.IsZero():
			scopes = append(scopes, p.buildWhereLessThenEndDate(filters.EndDate))
		}
	}

	// Sort
	if filters.SortField != "" {
		scopes = append(scopes, p.buildOrder(filters.SortField))
	} else {
		scopes = append(scopes, p.buildOrder("timestamp desc"))
	}

	return scopes
}

func (p *ProxySpeedLog) buildConditionsCount(filters ProxySpeedLogFilters) []func(*gorm.DB) *gorm.DB {
	scopes := make([]func(*gorm.DB) *gorm.DB, 0)

	// ProxyId filter
	if filters.ProxyId != "" {
		scopes = append(scopes, p.buildWhereProxyId(filters.ProxyId))
	}

	// Date filter
	if !filters.StartDate.IsZero() || !filters.EndDate.IsZero() {
		switch {
		case !filters.StartDate.IsZero() && !filters.EndDate.IsZero():
			scopes = append(scopes, p.buildWhereBetweenDates(filters.StartDate, filters.EndDate))
		case !filters.StartDate.IsZero():
			scopes = append(scopes, p.buildWhereMoreThenStartDate(filters.StartDate))
		case !filters.EndDate.IsZero():
			scopes = append(scopes, p.buildWhereLessThenEndDate(filters.EndDate))
		}
	}

	// Sort
	if filters.SortField != "" {
		scopes = append(scopes, p.buildOrder(filters.SortField))
	} else {
		scopes = append(scopes, p.buildOrder("timestamp desc"))
	}

	return scopes
}

func (p *ProxySpeedLog) List(filters ProxySpeedLogFilters, db *gorm.DB) ([]ProxySpeedLog, int64, error) {
	logs := []ProxySpeedLog{}
	limit := filters.PageSize
	offset := 0

	if filters.Page > 1 {
		offset = limit * (filters.Page - 1)
	}

	scopes := p.buildConditions(filters)
	err := db.Model(p).Scopes(scopes...).Limit(limit).Offset(offset).Find(&logs).Error
	var count int64
	err = db.Model(p).Scopes(p.buildConditionsCount(filters)...).Count(&count).Error
	return logs, count, err
}

type ProxyIPLog struct {
	Id         string    `json:"id"`
	ProxyId    string    `json:"proxy_id"`
	Timestamp  time.Time `json:"timestamp"`
	Ip         string    `json:"ip"`
	OldIp      string    `json:"old_ip"`
	Country    string    `json:"country"`
	OldCountry string    `json:"old_country"`
	ISP        string    `json:"isp"`
	OldISP     string    `json:"old_isp"`
}

func (i *ProxyIPLog) Save(db *gorm.DB) error {
	err := db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&i).Error
	return err
}

type ProxyIPLogFilters struct {
	ProxyId   string
	StartDate time.Time
	EndDate   time.Time
	Page      int
	PageSize  int
	SortField string
}

func (p *ProxyIPLog) buildWhereProxyId(proxyId string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("proxy_id = ?", proxyId)
	}
}

func (p *ProxyIPLog) buildWhereBetweenDates(startDate, endDate time.Time) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("timestamp BETWEEN ? AND ?", startDate, endDate)
	}
}

func (p *ProxyIPLog) buildWhereMoreThenStartDate(startDate time.Time) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("timestamp > ?", startDate)
	}
}

func (p *ProxyIPLog) buildWhereLessThenEndDate(endDate time.Time) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("timestamp < ?", endDate)
	}
}

func (p *ProxyIPLog) buildOrder(condition string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(condition)
	}
}

func (p *ProxyIPLog) buildConditions(filters ProxyIPLogFilters) []func(*gorm.DB) *gorm.DB {
	scopes := make([]func(*gorm.DB) *gorm.DB, 0)

	// ProxyId filter
	if filters.ProxyId != "" {
		scopes = append(scopes, p.buildWhereProxyId(filters.ProxyId))
	}

	// Date filter
	if !filters.StartDate.IsZero() || !filters.EndDate.IsZero() {
		switch {
		case !filters.StartDate.IsZero() && !filters.EndDate.IsZero():
			scopes = append(scopes, p.buildWhereBetweenDates(filters.StartDate, filters.EndDate))
		case !filters.StartDate.IsZero():
			scopes = append(scopes, p.buildWhereMoreThenStartDate(filters.StartDate))
		case !filters.EndDate.IsZero():
			scopes = append(scopes, p.buildWhereLessThenEndDate(filters.EndDate))
		}
	}

	// Sort
	if filters.SortField != "" {
		scopes = append(scopes, p.buildOrder(filters.SortField))
	} else {
		scopes = append(scopes, p.buildOrder("timestamp desc"))
	}

	return scopes
}

func (p *ProxyIPLog) buildConditionsCount(filters ProxyIPLogFilters) []func(*gorm.DB) *gorm.DB {
	scopes := make([]func(*gorm.DB) *gorm.DB, 0)

	// ProxyId filter
	if filters.ProxyId != "" {
		scopes = append(scopes, p.buildWhereProxyId(filters.ProxyId))
	}

	// Date filter
	if !filters.StartDate.IsZero() || !filters.EndDate.IsZero() {
		switch {
		case !filters.StartDate.IsZero() && !filters.EndDate.IsZero():
			scopes = append(scopes, p.buildWhereBetweenDates(filters.StartDate, filters.EndDate))
		case !filters.StartDate.IsZero():
			scopes = append(scopes, p.buildWhereMoreThenStartDate(filters.StartDate))
		case !filters.EndDate.IsZero():
			scopes = append(scopes, p.buildWhereLessThenEndDate(filters.EndDate))
		}
	}

	return scopes
}

func (p *ProxyIPLog) List(filters ProxyIPLogFilters, db *gorm.DB) ([]ProxyIPLog, int64, error) {
	logs := []ProxyIPLog{}
	limit := filters.PageSize
	offset := 0

	if filters.Page > 1 {
		offset = limit * (filters.Page - 1)
	}

	scopes := p.buildConditions(filters)
	err := db.Model(p).Scopes(scopes...).Limit(limit).Offset(offset).Find(&logs).Error

	var count int64
	err = db.Model(p).Scopes(p.buildConditionsCount(filters)...).Count(&count).Error
	return logs, count, err
}
