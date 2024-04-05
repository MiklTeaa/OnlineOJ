package log

import (
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"code-platform/pkg/osx"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "ts",
		NameKey:        "name",
		CallerKey:      "caller",
		FunctionKey:    "func",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000 +0800"),
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func newConfig() *zap.Config {
	logFilePath := os.Getenv("LOG_PATH")
	if logFilePath == "" {
		logFilePath = "stdout"
	} else {
		// zap 内置默认 Open 方法对Windows不大兼容，需要定制 RegisterSink
		if runtime.GOOS == "windows" {
			const scheme = "win"
			zap.RegisterSink(scheme, func(u *url.URL) (zap.Sink, error) {
				path := strings.TrimPrefix(u.Opaque, string(os.PathSeparator))
				return os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
			})
			logFilePath = filepath.Join("win:", logFilePath)
		}
	}

	if logFilePath != "stdout" {
		if err := osx.CreateFileIfNotExists(logFilePath); err != nil {
			panic(err)
		}
	}

	return &zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     newEncoderConfig(),
		OutputPaths:       []string{logFilePath},
		ErrorOutputPaths:  []string{logFilePath},
	}
}
