package main

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Record struct {
	// gorm.Model
	Id        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Provider  string    `json:"provider"`
	Product   string    `json:"product"`
	Weight    float64   `json:"weight"`
	Quantity  int       `json:"quantity"`
	Timestamp time.Time `json:"timestamp"`
	Comment   string    `json:"comment"`
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

func GetRecords(db *gorm.DB, fromDate string) ([]Record, error) {
	var records []Record
	query := db.Order("id desc")

	if fromDate != "" {
		query = query.Where("DATE(timestamp) = ?", fromDate)
	}

	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}

	return records, nil
}

func DeleteRecord(db *gorm.DB, id uint) error {
	return db.Delete(&Record{}, id).Error
}

func UpdateRecord(db *gorm.DB, id uint, quantity int, comment string) error {
	updates := map[string]interface{}{}
	if quantity != 0 {
		updates["quantity"] = quantity
	}
	if comment != "" {
		updates["comment"] = comment
	}

	return db.Model(&Record{}).Where("id = ?", id).Updates(updates).Error
}
