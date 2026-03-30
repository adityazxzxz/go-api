package helpers

import (
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Global logger
var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func SetupLogging() {

	// Pastikan folder logs ada
	os.MkdirAll("logs", os.ModePerm)

	// Access log file
	accessLog := &lumberjack.Logger{
		Filename:   "logs/access.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
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

	gin.DefaultWriter = io.MultiWriter(accessLog, os.Stdout)
	gin.DefaultErrorWriter = io.MultiWriter(errorLog, os.Stderr)

	InfoLogger = log.New(io.MultiWriter(accessLog, os.Stdout), "INFO: ", log.LstdFlags)
	ErrorLogger = log.New(io.MultiWriter(errorLog, os.Stderr), "ERROR: ", log.LstdFlags)

	log.SetOutput(io.MultiWriter(accessLog, os.Stdout))
}
