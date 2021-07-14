package log

import "log"

// DefaultLogger default logger based on std logger
type DefaultLogger struct{}

// IsProduction this basic default logger used only for testing
func (l *DefaultLogger) IsProduction() bool {
	return false
}

// Rawf log
func (l *DefaultLogger) Rawf(rawMessage []byte, format string, v ...interface{}) {
	log.Printf("[ RAW "+string(rawMessage)+" ] "+format, v...)
}

// Debugf log info
func (l *DefaultLogger) Debugf(format string, v ...interface{}) {
	log.Printf("[ DEBUG ] "+format, v...)
}

// Infof log info
func (l *DefaultLogger) Infof(format string, v ...interface{}) {
	log.Printf("[ INFO ] "+format, v...)
}

// Errorf log error
func (l *DefaultLogger) Errorf(err error, format string, v ...interface{}) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	log.Printf("[ ERROR : "+errMsg+" ]"+format, v...)
}

// Fatalf log fatal
func (l *DefaultLogger) Fatalf(err error, format string, v ...interface{}) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	log.Fatalf("[ FATAL : "+errMsg+" ]"+format, v...)
}
