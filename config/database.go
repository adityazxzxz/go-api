package config

import (
	"fmt"
	"go-api/helpers"
	"log"
	"os"
	"strconv"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupLogFile(path string) *os.File {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		helpers.ErrorLogger.Panic("Failed to open log file:", err)
		panic(err)
	}
	return f
}

func DBInit() (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	gormLogFile := setupLogFile("logs/database.log")

	newLogger := logger.New(
		log.New(gormLogFile, "", log.LstdFlags),
		logger.Config{
			LogLevel: logger.Error, // hanya error
		},
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Silent),
		Logger: newLogger,
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
