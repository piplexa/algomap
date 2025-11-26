package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// Init инициализирует глобальный логгер
func Init(level string) error {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapLevel),
		Encoding:         "console", // "console" для читаемости, "json" для продакшена
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalColorLevelEncoder, // 🎨 Цветной лог!
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}

	var err error
	Log, err = config.Build()
	if err != nil {
		return err
	}

	return nil
}

// Sync сбрасывает буферы логгера (вызывать перед завершением программы)
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// Debug логирует debug сообщение
func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

// Info логирует info сообщение
func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

// Warn логирует warning сообщение
func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

// Error логирует error сообщение
func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

// Fatal логирует fatal сообщение и завершает программу
func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}