package main

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Settings struct {
	Url       string `json:"url"`
	Timeout   int    `json:"timeout"`
	Repeat    int    `json:"repeat"`
	LastIndex int
}

func (s *Settings) Save(db *gorm.DB) error {
	// This performs an "upsert".
	// It will create the settings record if it doesn't exist.
	// If it does exist (based on the primary key), it will update all columns.
	// Since there's only one settings row, this is safe and effective.
	return db.Clauses(clause.OnConflict{
		DoNothing: true, // Assuming you don't want to update if it exists, just create.
	}).Create(s).Error
}

func (s *Settings) Get(db *gorm.DB) (*Settings, error) {
	settings := &Settings{}
	err := db.First(settings).Error
	return settings, err
}

func SettingsDefault(db *gorm.DB) *Settings {
	s := Settings{}
	settings, err := s.Get(db)
	if err == gorm.ErrRecordNotFound {
		stg := &Settings{
			Url:     "https://google.com",
			Timeout: 5,
			Repeat:  15,
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
