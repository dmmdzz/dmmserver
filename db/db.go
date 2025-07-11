// internal/db/db.go
package db

import (

	"fmt"
	"log"
	"dmmserver/conf"
	"dmmserver/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	log.Println("Initializing database connection...")
	var err error
	c := conf.Conf.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Name)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Auto-migrating database tables...")
	err = DB.AutoMigrate(
		&model.BanDeviceID{},
		&model.BanIP{},
		&model.BanRealDeviceID{},
		&model.PlayerData{},
		&model.ServerSettings{},
		&model.PlayerInfo{},
		&model.BanDeviceInfo{},
	)
	if err != nil {
		log.Fatalf("Failed to auto-migrate tables: %v", err)
	}

	log.Println("Database initialization and migration complete.")
}