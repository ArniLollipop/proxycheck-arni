package main

import (
	"gorm.io/gorm"
)

type Settings struct {
	ID                 uint   `gorm:"primaryKey" json:"-"` // Добавляем ID как первичный ключ
	Url                string `json:"url"`
	Timeout            int    `json:"timeout"`
	CheckIPInterval    int    `json:"checkIPInterval"`
	SpeedCheckInterval int    `json:"speedCheckInterval"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	SkipSSLVerify      bool   `json:"skipSSLVerify"` // Allow configuring SSL verification

	// Notification settings
	TelegramEnabled      bool   `json:"telegramEnabled"`
	TelegramToken        string `json:"telegramToken"`
	TelegramChatID       string `json:"telegramChatID"`
	NotifyOnDown         bool   `json:"notifyOnDown"`         // Notify when proxy goes down
	NotifyOnRecovery     bool   `json:"notifyOnRecovery"`     // Notify when proxy recovers
	NotifyOnIPChange     bool   `json:"notifyOnIPChange"`     // Notify when IP changes
	NotifyOnIPStuck      bool   `json:"notifyOnIPStuck"`      // Notify when IP is stuck >24h
	NotifyOnLowSpeed     bool   `json:"notifyOnLowSpeed"`     // Notify when speed is low
	LowSpeedThreshold    int    `json:"lowSpeedThreshold"`    // Mbps threshold for low speed
	NotifyDailySummary   bool   `json:"notifyDailySummary"`   // Send daily summary
	DailySummaryTime     string `json:"dailySummaryTime"`     // Time for daily summary (HH:MM format)
}

func (s *Settings) Save(db *gorm.DB) error {
	s.ID = 1
	return db.Save(s).Error
}

func (s *Settings) Get(db *gorm.DB) (*Settings, error) {
	settings := &Settings{}
	// Ищем запись с ID=1
	err := db.First(settings, 1).Error
	return settings, err
}

func SettingsDefault(db *gorm.DB) *Settings {
	s := Settings{}
	settings, err := s.Get(db)
	if settings.Username == "" {
		settings.Username = "default_username"
		settings.Password = "default_password"
		settings.Save(db)
	}

	if err == gorm.ErrRecordNotFound {
		stg := &Settings{
			ID:                 1,
			Url:                "https://google.com",
			Timeout:            5,
			CheckIPInterval:    5,
			SpeedCheckInterval: 15,
			Username:           "default_username",
			Password:           "default_password",
			SkipSSLVerify:      true, // Default to true for backward compatibility
			// Notification defaults
			TelegramEnabled:    false,
			NotifyOnDown:       true,
			NotifyOnRecovery:   true,
			NotifyOnIPChange:   false,
			NotifyOnIPStuck:    true,
			NotifyOnLowSpeed:   false,
			LowSpeedThreshold:  10, // 10 Mbps
			NotifyDailySummary: false,
			DailySummaryTime:   "09:00",
		}
		err := stg.Save(db)
		if err != nil {
			panic(err)
		}

	} else if err != nil {
		panic(err)
	}

	return settings
}
