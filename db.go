package main

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Record struct {
	gorm.Model
	Id        int
	Provider  string
	Product   string
	Weigth    float32
	Quantity  int8
	Timestamp time.Time
}

func GetDbSession() *gorm.DB {
	session, err := gorm.Open(sqlite.Open("bouffeaction.db"), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	session.AutoMigrate(&Record{})

	return session

}
