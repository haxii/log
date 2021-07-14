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

var GlobalLogger = MakeSimpleZeroLogger()

func Rawf(msg []byte, format string, v ...interface{})     { GlobalLogger.Rawf(msg, format, v...) }
func Debugf(format string, args ...interface{})            { GlobalLogger.Debugf(format, args...) }
func Infof(format string, args ...interface{})             { GlobalLogger.Infof(format, args...) }
func Errorf(err error, format string, args ...interface{}) { GlobalLogger.Errorf(err, format, args...) }
func Fatalf(err error, format string, args ...interface{}) { GlobalLogger.Fatalf(err, format, args...) }
