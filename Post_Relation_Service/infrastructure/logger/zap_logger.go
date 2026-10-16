package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	log *zap.Logger
}

func NewZapLogger() (Logger, error) {

	config := zap.NewProductionConfig()

	// 🔹 Use readable timestamp
	config.EncoderConfig.TimeKey = "ts"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 🔹 Short caller (file:line)
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// 🔹 Optional: disable stacktrace for error logs
	config.DisableStacktrace = true

	zapLogger, err := config.Build(
		zap.AddCaller(),      // include caller info
		zap.AddCallerSkip(1), // skip wrapper layer
	)
	if err != nil {
		return nil, err
	}

	return &ZapLogger{
		log: zapLogger,
	}, nil
}

func toZapFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zapFields = append(zapFields, zap.Any(f.Key, f.Value))
	}
	return zapFields
}

func (l *ZapLogger) Info(msg string, fields ...Field) {
	l.log.Info(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Debug(msg string, fields ...Field) {
	l.log.Debug(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Warn(msg string, fields ...Field) {
	l.log.Warn(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Error(msg string, fields ...Field) {
	l.log.Error(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Fatal(msg string, fields ...Field) {
	l.log.Fatal(msg, toZapFields(fields)...)
}

func (l *ZapLogger) With(fields ...Field) Logger {
	return &ZapLogger{
		log: l.log.With(toZapFields(fields)...),
	}
}

func (l *ZapLogger) Sync() error {
	return l.log.Sync()
}
