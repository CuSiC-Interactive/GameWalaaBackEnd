package utils

import (
    "log"
    "os"
    "path/filepath"
    "runtime"
    "time"
)

var (
    InfoLogger  *log.Logger
    ErrorLogger *log.Logger
    logFile    *os.File
)

// InitLogger initializes the logger with a file in the logs directory
func InitLogger() error {
    // Create logs directory if it doesn't exist
    logsDir := "logs"
    if err := os.MkdirAll(logsDir, 0755); err != nil {
        return err
    }

    // Create or append to log file with current date
    currentTime := time.Now()
    logFileName := filepath.Join(logsDir, currentTime.Format("2006-01-02")+".log")
    file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }

    logFile = file
    InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime)
    ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime)
    
    return nil
}

// CloseLogger closes the log file
func CloseLogger() {
    if logFile != nil {
        logFile.Close()
    }
}

// LogInfo logs information messages
func LogInfo(format string, v ...interface{}) {
    if InfoLogger != nil {
        // Get caller file and line
        _, file, line, ok := runtime.Caller(1)
        if ok {
            // Extract just the file name from the full path
            file = filepath.Base(file)
            InfoLogger.Printf("%s:%d: "+format, append([]interface{}{file, line}, v...)...)
        } else {
            InfoLogger.Printf(format, v...)
        }
    }
}

// LogError logs error messages
func LogError(format string, v ...interface{}) {
    if ErrorLogger != nil {
        // Get caller file and line
        _, file, line, ok := runtime.Caller(1)
        if ok {
            // Extract just the file name from the full path
            file = filepath.Base(file)
            ErrorLogger.Printf("%s:%d: "+format, append([]interface{}{file, line}, v...)...)
        } else {
            ErrorLogger.Printf(format, v...)
        }
    }
}
