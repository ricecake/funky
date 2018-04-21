package datastore

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
)

type StorageRecord struct {
	Id    string `gorm:"column:id"`
	Owner string `gorm:"column:owner"`
	Scope string `gorm:"column:scope"`
	Type  string `gorm:"column:type"`
	Key   string `gorm:"column:key"`
	Value string `gorm:"column:value"`
}

func (StorageRecord) TableName() string {
	return "storage"
}

func (StorageRecord) Load() error {
	return nil
}

var db *gorm.DB

func Connect() (*gorm.DB, error) {
	var err error
	if db == nil {
		username := viper.GetString("db.username")
		password := viper.GetString("db.password")
		host := viper.GetString("db.host")
		database := viper.GetString("db.database")
		db, err = gorm.Open("postgres", fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=disable", username, password, host, database))
		db.LogMode(true)
	}

	return db, err
}

func SetDb(newDb *gorm.DB) {
	db = newDb
}

func GetDb() *gorm.DB {
	if db == nil {
		if _, err := Connect(); err != nil {
			panic(err)
		}
	}
	return db
}

func init() {
	return
}
