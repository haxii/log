package log

//Logger proxy logger, used for logging proxy info and errors
type Logger interface {
	Debug(who, format string, v ...interface{})
	Info(who, format string, v ...interface{})
	Error(who string, err error, format string, v ...interface{})
	Fatal(who string, err error, format string, v ...interface{})
}
