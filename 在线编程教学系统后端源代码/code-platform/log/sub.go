package log

func (l *Logger) Sub(name string) *Logger {
	return &Logger{
		logger: l.logger.Named(name),
	}
}

func Sub(name string) *Logger {
	return defaultLogger.Sub(name)
}
