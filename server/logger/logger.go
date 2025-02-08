package logger

import (
	"go.uber.org/zap"
	"sync"
)

var (
	Logger *zap.Logger
	once   sync.Once // Ensures that the logger is only initialized once
)

func Init() {
	once.Do(func() {
		var err error
		Logger, err = zap.NewProduction()
		if err != nil {
			panic(err)
		}
	})
}
