package log

// Logger proxy logger, used for logging proxy info and errors
type Logger interface {
	Rawf(rawMessage []byte, format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(err error, format string, args ...interface{})
	Fatalf(err error, format string, args ...interface{})
}

var GlobalLogger = MakeSimpleZeroLogger()

func Rawf(msg []byte, format string, v ...interface{}) {
	GlobalLogger.rawf(defaultCallSkip, msg, format, v...)
}
func Debugf(format string, args ...interface{}) {
	GlobalLogger.debugf(defaultCallSkip, format, args...)
}
func Infof(format string, args ...interface{}) {
	GlobalLogger.infof(defaultCallSkip, format, args...)
}
func Errorf(err error, format string, args ...interface{}) {
	GlobalLogger.errorf(defaultCallSkip, err, format, args...)
}
func Fatalf(err error, format string, args ...interface{}) {
	GlobalLogger.fatalf(defaultCallSkip, err, format, args...)
}
