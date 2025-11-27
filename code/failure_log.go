package main

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProxyFailureLog represents a failure event for a proxy
type ProxyFailureLog struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	ProxyID   string    `json:"proxy_id" gorm:"index"`
	Timestamp time.Time `json:"timestamp" gorm:"index"`
	ErrorType string    `json:"error_type"` // "ping_failed", "speed_check_failed", "ip_check_failed"
	ErrorMsg  string    `json:"error_msg"`
	Latency   int       `json:"latency"` // Last known latency before failure
}

// TableName specifies the table name for ProxyFailureLog
func (ProxyFailureLog) TableName() string {
	return "proxy_failure_logs"
}

// Save creates or updates a failure log
func (f *ProxyFailureLog) Save(db *gorm.DB) error {
	return db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(f).Error
}

// ProxyFailureLogFilters represents filters for querying failure logs
type ProxyFailureLogFilters struct {
	ProxyID   string
	ErrorType string
	StartDate time.Time
	EndDate   time.Time
	Page      int
	PageSize  int
	SortField string
}

// buildWhereProxyID builds a WHERE clause for proxy_id
func (f *ProxyFailureLog) buildWhereProxyID(proxyID string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("proxy_id = ?", proxyID)
	}
}

// buildWhereErrorType builds a WHERE clause for error_type
func (f *ProxyFailureLog) buildWhereErrorType(errorType string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("error_type = ?", errorType)
	}
}

// buildWhereBetweenDates builds a WHERE clause for date range
func (f *ProxyFailureLog) buildWhereBetweenDates(startDate, endDate time.Time) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("timestamp BETWEEN ? AND ?", startDate, endDate)
	}
}

// buildWhereMoreThenStartDate builds a WHERE clause for dates after start_date
func (f *ProxyFailureLog) buildWhereMoreThenStartDate(startDate time.Time) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("timestamp > ?", startDate)
	}
}

// buildWhereLessThenEndDate builds a WHERE clause for dates before end_date
func (f *ProxyFailureLog) buildWhereLessThenEndDate(endDate time.Time) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("timestamp < ?", endDate)
	}
}

// buildOrder builds an ORDER BY clause
func (f *ProxyFailureLog) buildOrder(condition string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(condition)
	}
}

// buildConditions builds all conditions for querying
func (f *ProxyFailureLog) buildConditions(filters ProxyFailureLogFilters) []func(*gorm.DB) *gorm.DB {
	scopes := make([]func(*gorm.DB) *gorm.DB, 0)

	// ProxyID filter
	if filters.ProxyID != "" {
		scopes = append(scopes, f.buildWhereProxyID(filters.ProxyID))
	}

	// ErrorType filter
	if filters.ErrorType != "" {
		scopes = append(scopes, f.buildWhereErrorType(filters.ErrorType))
	}

	// Date filter
	if !filters.StartDate.IsZero() || !filters.EndDate.IsZero() {
		switch {
		case !filters.StartDate.IsZero() && !filters.EndDate.IsZero():
			scopes = append(scopes, f.buildWhereBetweenDates(filters.StartDate, filters.EndDate))
		case !filters.StartDate.IsZero():
			scopes = append(scopes, f.buildWhereMoreThenStartDate(filters.StartDate))
		case !filters.EndDate.IsZero():
			scopes = append(scopes, f.buildWhereLessThenEndDate(filters.EndDate))
		}
	}

	// Sort
	if filters.SortField != "" {
		scopes = append(scopes, f.buildOrder(filters.SortField))
	} else {
		scopes = append(scopes, f.buildOrder("timestamp desc"))
	}

	return scopes
}

// buildConditionsCount builds conditions for counting (without sorting)
func (f *ProxyFailureLog) buildConditionsCount(filters ProxyFailureLogFilters) []func(*gorm.DB) *gorm.DB {
	scopes := make([]func(*gorm.DB) *gorm.DB, 0)

	// ProxyID filter
	if filters.ProxyID != "" {
		scopes = append(scopes, f.buildWhereProxyID(filters.ProxyID))
	}

	// ErrorType filter
	if filters.ErrorType != "" {
		scopes = append(scopes, f.buildWhereErrorType(filters.ErrorType))
	}

	// Date filter
	if !filters.StartDate.IsZero() || !filters.EndDate.IsZero() {
		switch {
		case !filters.StartDate.IsZero() && !filters.EndDate.IsZero():
			scopes = append(scopes, f.buildWhereBetweenDates(filters.StartDate, filters.EndDate))
		case !filters.StartDate.IsZero():
			scopes = append(scopes, f.buildWhereMoreThenStartDate(filters.StartDate))
		case !filters.EndDate.IsZero():
			scopes = append(scopes, f.buildWhereLessThenEndDate(filters.EndDate))
		}
	}

	return scopes
}

// List retrieves failure logs with pagination
func (f *ProxyFailureLog) List(filters ProxyFailureLogFilters, db *gorm.DB) ([]ProxyFailureLog, int64, error) {
	logs := []ProxyFailureLog{}
	limit := filters.PageSize
	offset := 0

	if filters.Page > 1 {
		offset = limit * (filters.Page - 1)
	}

	scopes := f.buildConditions(filters)
	countScopes := f.buildConditionsCount(filters)

	var count int64
	// Count first
	if err := db.Model(f).Scopes(countScopes...).Count(&count).Error; err != nil {
		return logs, 0, err
	}

	// Then fetch data
	err := db.Model(f).Scopes(scopes...).Limit(limit).Offset(offset).Find(&logs).Error
	return logs, count, err
}

// GetStats returns failure statistics for a proxy
type FailureStats struct {
	TotalFailures     int64   `json:"total_failures"`
	PingFailures      int64   `json:"ping_failures"`
	SpeedFailures     int64   `json:"speed_failures"`
	IPCheckFailures   int64   `json:"ip_check_failures"`
	LastFailure       *time.Time `json:"last_failure"`
	FailureRate       float64 `json:"failure_rate"` // Failures per day
}

// GetFailureStats calculates failure statistics for a proxy
func GetFailureStats(db *gorm.DB, proxyID string, days int) (*FailureStats, error) {
	stats := &FailureStats{}

	startDate := time.Now().AddDate(0, 0, -days)

	// Total failures
	db.Model(&ProxyFailureLog{}).
		Where("proxy_id = ? AND timestamp > ?", proxyID, startDate).
		Count(&stats.TotalFailures)

	// Ping failures
	db.Model(&ProxyFailureLog{}).
		Where("proxy_id = ? AND timestamp > ? AND error_type = ?", proxyID, startDate, "ping_failed").
		Count(&stats.PingFailures)

	// Speed check failures
	db.Model(&ProxyFailureLog{}).
		Where("proxy_id = ? AND timestamp > ? AND error_type = ?", proxyID, startDate, "speed_check_failed").
		Count(&stats.SpeedFailures)

	// IP check failures
	db.Model(&ProxyFailureLog{}).
		Where("proxy_id = ? AND timestamp > ? AND error_type = ?", proxyID, startDate, "ip_check_failed").
		Count(&stats.IPCheckFailures)

	// Last failure
	var lastLog ProxyFailureLog
	err := db.Model(&ProxyFailureLog{}).
		Where("proxy_id = ?", proxyID).
		Order("timestamp desc").
		Limit(1).
		First(&lastLog).Error

	if err == nil {
		stats.LastFailure = &lastLog.Timestamp
	}

	// Calculate failure rate (failures per day)
	if days > 0 {
		stats.FailureRate = float64(stats.TotalFailures) / float64(days)
	}

	return stats, nil
}
