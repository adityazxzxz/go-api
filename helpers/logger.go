package helpers

import (
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
)

func SetupLogging() {

	// Pastikan folder logs ada
	os.MkdirAll("logs", os.ModePerm)

	// Access log file
	accessLog := &lumberjack.Logger{
		Filename:   "logs/access.log",
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     30, // hari
		Compress:   true,
	}

	// Error log file
	errorLog := &lumberjack.Logger{
		Filename:   "logs/error.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}

	// Set output
	gin.DefaultWriter = io.MultiWriter(accessLog, os.Stdout)
	gin.DefaultErrorWriter = io.MultiWriter(errorLog, os.Stderr)
}
