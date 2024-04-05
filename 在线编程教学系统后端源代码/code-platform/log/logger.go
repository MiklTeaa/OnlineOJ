package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	logger *zap.Logger
}

var defaultLogger *Logger

func init() {
	logger, err := newConfig().Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
	logger = logger.Named("root")
	defaultLogger = &Logger{
		logger: logger,
	}
}

func Errorf(err error, msg string, format ...interface{}) {
	defaultLogger.Errorf(err, msg, format...)
}

func Error(err error, msg string) {
	defaultLogger.Error(err, msg)
}

func Warnf(msg string, format ...interface{}) {
	defaultLogger.Warnf(msg, format...)
}

func Warn(msg string) {
	defaultLogger.Warn(msg)
}

func Infof(msg string, format ...interface{}) {
	defaultLogger.Infof(msg, format...)
}

func Info(msg string) {
	defaultLogger.Info(msg)
}

func Debugf(msg string, format ...interface{}) {
	defaultLogger.Debugf(msg, format...)
}

func Debug(msg string) {
	defaultLogger.Debug(msg)
}

func (l *Logger) Errorf(err error, msg string, format ...interface{}) {
	if len(format) > 0 {
		msg = fmt.Sprintf(msg, format...)
	}

	var errString string
	if err != nil {
		errString = err.Error()
	}
	l.logger.With(zapcore.Field{
		Key:    "err",
		Type:   zapcore.StringType,
		String: errString,
	}).Error(msg)
}

func (l *Logger) Error(err error, msg string) {
	var errString string
	if err != nil {
		errString = err.Error()
	}
	l.logger.With(zapcore.Field{
		Key:    "err",
		Type:   zapcore.StringType,
		String: errString,
	}).Error(msg)
}

func (l *Logger) Warnf(msg string, format ...interface{}) {
	if len(format) > 0 {
		msg = fmt.Sprintf(msg, format...)
	}
	l.logger.Warn(msg)
}

func (l *Logger) Warn(msg string) {
	l.logger.Warn(msg)
}

func (l *Logger) Infof(msg string, format ...interface{}) {
	if len(format) > 0 {
		msg = fmt.Sprintf(msg, format...)
	}
	l.logger.Info(msg)
}

func (l *Logger) Info(msg string) {
	l.logger.Info(msg)
}

func (l *Logger) Debugf(msg string, format ...interface{}) {
	if len(format) > 0 {
		msg = fmt.Sprintf(msg, format...)
	}
	l.logger.Debug(msg)
}

func (l *Logger) Debug(msg string) {
	l.logger.Debug(msg)
}
