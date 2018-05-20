package log

import "log"

// DefaultLogger default logger based on std logger
type DefaultLogger struct{}

// Debug log info
func (l *DefaultLogger) Debug(who, format string, v ...interface{}) {
	log.Printf("[ DEBUG "+who+" ] "+format, v...)
}

// Info log info
func (l *DefaultLogger) Info(who, format string, v ...interface{}) {
	log.Printf("[ INFO "+who+" ] "+format, v...)
}

// Error log error
func (l *DefaultLogger) Error(who string, err error, format string, v ...interface{}) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	log.Printf("[ ERROR "+who+" : "+errMsg+" ]"+format, v...)
}

// Fatal log fatal
func (l *DefaultLogger) Fatal(who string, err error, format string, v ...interface{}) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	log.Fatalf("[ FATAL "+who+" : "+errMsg+" ]"+format, v...)
}
