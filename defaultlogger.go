package log

import "log"

// DefaultLogger default logger based on std logger
type DefaultLogger struct{}

// Debug log info
func (l *DefaultLogger) Debug(format string, v ...interface{}) {
	log.Printf("[ DEBUG ] "+format, v...)
}

// Info log info
func (l *DefaultLogger) Info(format string, v ...interface{}) {
	log.Printf("[ INFO ] "+format, v...)
}

// Error log error
func (l *DefaultLogger) Error(err error, format string, v ...interface{}) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	log.Printf("[ ERROR : "+errMsg+" ]"+format, v...)
}

// Fatal log fatal
func (l *DefaultLogger) Fatal(err error, format string, v ...interface{}) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	log.Fatalf("[ FATAL : "+errMsg+" ]"+format, v...)
}
