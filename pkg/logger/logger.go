package logger

import (
	"github.com/olusolaa/github-monitor/pkg/errors"
	"log"
	"os"
)

// Define color codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	White  = "\033[37m"
)

// Logger instance
var logger *log.Logger

// InitLogger initializes the logger with standard settings
func InitLogger() {
	logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
}

// Colorize applies the given color to the text
func Colorize(color, text string) string {
	return color + text + Reset
}

// LogError logs an error message with color based on severity
func LogError(err error) {
	if err != nil && logger != nil {
		color := White
		if customErr, ok := err.(*errors.CustomError); ok {
			switch customErr.Severity {
			case errors.Critical:
				color = Red
			case errors.Warning:
				color = Yellow
			case errors.Info:
				color = Green
			}
			logger.Println(Colorize(color, "ERROR: "+customErr.Error()))
		} else {
			logger.Println(Colorize(color, "ERROR: "+err.Error()))
		}
	}
}

// LogInfo logs an informational message in green
func LogInfo(msg string) {
	if logger != nil {
		logger.Println(Colorize(Green, "INFO: "+msg))
	}
}

// LogWarning logs a warning message in yellow
func LogWarning(msg string) {
	if logger != nil {
		logger.Println(Colorize(Yellow, "WARNING: "+msg))
	}
}

// LogDebug logs a debug message in blue
func LogDebug(msg string) {
	if logger != nil {
		logger.Println(Colorize(Blue, "DEBUG: "+msg))
	}
}
