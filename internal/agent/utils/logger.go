package utils

import (
	"go.uber.org/zap"
	"log"
)

func InitLogger() *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't get new logger in utils/loggers.go: %v", err)
	}
	return logger
}
