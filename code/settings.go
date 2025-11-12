package main

import (
	"gorm.io/gorm"
)

type Settings struct {
	ID                 uint   `gorm:"primaryKey" json:"-"` // Добавляем ID как первичный ключ
	Url                string `json:"url"`
	Timeout            int    `json:"timeout"`
	Repeat             int    `json:"repeat"`
	CheckIPInterval    int    `json:"checkIPInterval"`
	SpeedCheckInterval int    `json:"speedCheckInterval"`
	Username           string `json:"username"`
	Password           string `json:"password"`
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
			Repeat:             15,
			CheckIPInterval:    15,
			SpeedCheckInterval: 1440,
			Username:           "default_username",
			Password:           "default_password",
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
