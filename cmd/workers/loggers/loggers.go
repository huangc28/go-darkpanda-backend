package loggers

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	errLogger  = log.New()
	infoLogger = log.New()
)

func GetErrorLogger() *log.Logger {
	return errLogger
}

func GetInfoLogger() *log.Logger {
	return infoLogger
}

func InitErrLogger(errLogPath, logName string) {
	if err := os.MkdirAll(errLogPath, os.ModePerm); err != nil {
		log.Fatalf("failed to create file: %v", err)
	}

	errLogger.SetFormatter(&log.JSONFormatter{})

	file, err := os.OpenFile(
		fmt.Sprintf(
			"%s/%s_%s_error.log",
			errLogPath,
			time.Now().Format("01-02-2006"),
			logName,
		),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0755,
	)

	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	errLogger.Out = file
	errLogger.SetLevel(log.ErrorLevel)
}

func InitInfoLogger(infoLogPath, logName string) {
	if err := os.MkdirAll(infoLogPath, os.ModePerm); err != nil {
		log.Fatalf("failed to create file: %v", err)
	}

	errLogger.SetFormatter(&log.JSONFormatter{})

	file, err := os.OpenFile(
		fmt.Sprintf(
			"%s/%s_%s_info.log",
			infoLogPath,
			time.Now().Format("01-02-2006"),
			logName,
		),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0755,
	)

	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	infoLogger.Out = file
	infoLogger.SetLevel(log.InfoLevel)
}
