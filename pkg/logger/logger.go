package logger

import (
	"log"
	"os"
)

var (
	logger *log.Logger
)

func InitLogger() {
	logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
}

func LogError(err error) {
	if err != nil && logger != nil {
		logger.Println("ERROR:", err)
	}
}

func LogInfo(msg string) {
	if logger != nil {
		logger.Println("INFO:", msg)
	}
}
