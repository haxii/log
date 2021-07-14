package log

// Logger proxy logger, used for logging proxy info and errors
type Logger interface {
	IsProduction() bool
	Rawf(rawMessage []byte, format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(err error, format string, args ...interface{})
	Fatalf(err error, format string, args ...interface{})
}
