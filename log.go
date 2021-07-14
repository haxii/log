package log

// Logger proxy logger, used for logging proxy info and errors
type Logger interface {
	IsProduction() bool
	Raw(rawMessage []byte, format string, v ...interface{})
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Error(err error, format string, v ...interface{})
	Fatal(err error, format string, v ...interface{})
}
