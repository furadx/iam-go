package log

import (
	"fmt"
	"log"
)

// Info logs an info message.
func Info(msg string) {
	log.Println("[INFO]", msg)
}

// Infof logs a formatted info message.
func Infof(format string, args ...interface{}) {
	log.Println("[INFO]", fmt.Sprintf(format, args...))
}

// Error logs an error message.
func Error(msg string) {
	log.Println("[ERROR]", msg)
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...interface{}) {
	log.Println("[ERROR]", fmt.Sprintf(format, args...))
}

// Warn logs a warning message.
func Warn(msg string) {
	log.Println("[WARN]", msg)
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...interface{}) {
	log.Println("[WARN]", fmt.Sprintf(format, args...))
}

// Debug logs a debug message.
func Debug(msg string) {
	log.Println("[DEBUG]", msg)
}

// Debugf logs a formatted debug message.
func Debugf(format string, args ...interface{}) {
	log.Println("[DEBUG]", fmt.Sprintf(format, args...))
}
