package logger

import (
	"github.com/sirupsen/logrus"
)

type LogrusLogger struct {
	log *logrus.Logger
}

func NewLogrusLogger() *LogrusLogger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	return &LogrusLogger{log: log}
}

func (l *LogrusLogger) Info(msg string, fields map[string]interface{}) {
	l.log.WithFields(logrus.Fields(fields)).Info(msg)
}

func (l *LogrusLogger) Error(msg string, fields map[string]interface{}) {
	l.log.WithFields(logrus.Fields(fields)).Error(msg)
}

func (l *LogrusLogger) Debug(msg string, fields map[string]interface{}) {
	l.log.WithFields(logrus.Fields(fields)).Debug(msg)
}