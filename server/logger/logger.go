package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"sync"
)

var (
	Logger *zap.Logger
	once   sync.Once // Ensures that the logger is only initialized once
)

func Init() {
	once.Do(func() {
		writeSyncer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   "storage/logs/server.log",
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28,   // days
			Compress:   true, // disabled by default
		})

		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(writeSyncer), zapcore.AddSync(zapcore.Lock(os.Stdout))),
			zap.InfoLevel,
		)

		Logger = zap.New(core)
	})
}
