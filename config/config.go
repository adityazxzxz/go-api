package config

import (
	"fmt"
	"os"
	"strconv"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func DBInit() (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		// Logger: newLogger,
	})
	if err != nil {
		panic("cannot connect database")
	}

	IdleConn, _ := strconv.Atoi(os.Getenv("DB_IDLE_CONN"))
	MaxOpenConn, _ := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONN"))

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(IdleConn)
	sqlDB.SetMaxOpenConns(MaxOpenConn)

	return db, nil
}
