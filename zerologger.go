package log

import (
	"io"
	"os"
	"time"

	"encoding/json"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

// ZeroLogger implemented logger using zerolog
type ZeroLogger struct {
	logger  zerolog.Logger
	logFile *os.File

	callerLevel zerolog.Level
}

// GetZeroLogger returns the zero logger instance for advanced usage
func (l *ZeroLogger) GetZeroLogger() zerolog.Logger {
	return l.logger
}

const defaultCallSkip = 3

// Rawf implements raw logger interface
func (l *ZeroLogger) Rawf(rawMessage []byte, format string, v ...interface{}) {
	l.rawf(defaultCallSkip, rawMessage, format, v...)
}
func (l *ZeroLogger) rawf(callSkip int, rawMessage []byte, format string, v ...interface{}) {
	event := l.logger.WithLevel(zerolog.NoLevel).Caller(callSkip)
	if json.Valid(rawMessage) {
		rawMessageInJSON := json.RawMessage(rawMessage)
		event.Interface("raw", rawMessageInJSON)
	} else {
		event.Bytes("raw", rawMessage)
	}
	if len(v) == 0 {
		event.Msg(format)
	} else {
		event.Msgf(format, v...)
	}
}

// Debugf implements debug logger interface
func (l *ZeroLogger) Debugf(format string, v ...interface{}) {
	l.debugf(defaultCallSkip, format, v...)
}
func (l *ZeroLogger) debugf(callSkip int, format string, v ...interface{}) {
	eventf(l.logger.Debug(), zerolog.DebugLevel >= l.callerLevel, false, callSkip, format, v...)
}

// Infof implements info logger interface
func (l *ZeroLogger) Infof(format string, v ...interface{}) {
	l.infof(defaultCallSkip, format, v...)
}
func (l *ZeroLogger) infof(callSkip int, format string, v ...interface{}) {
	eventf(l.logger.Info(), zerolog.InfoLevel >= l.callerLevel, false, callSkip, format, v...)
}

// Errorf implements error logger interface
func (l *ZeroLogger) Errorf(err error, format string, v ...interface{}) {
	l.errorf(defaultCallSkip, err, format, v...)
}
func (l *ZeroLogger) errorf(callSkip int, err error, format string, v ...interface{}) {
	withCaller := zerolog.ErrorLevel >= l.callerLevel
	eventf(l.logger.Error().Err(err), withCaller, withCaller, callSkip, format, v...)
}

// Fatalf make a fatal return
func (l *ZeroLogger) Fatalf(err error, format string, v ...interface{}) {
	l.fatalf(defaultCallSkip, err, format, v...)
}
func (l *ZeroLogger) fatalf(callSkip int, err error, format string, v ...interface{}) {
	eventf(l.logger.Panic().Err(err), true, true, callSkip, format, v...)
}

func eventf(event *zerolog.Event, withCaller, withStack bool, callSkip int, format string, v ...interface{}) {
	if withCaller {
		event = event.Caller(callSkip)
	}
	if withStack {
		event = event.Stack()
	}
	if len(v) == 0 {
		event.Msg(format)
	} else {
		event.Msgf(format, v...)
	}
}

// LoggingConfig helper for a logging destination
type LoggingConfig struct {
	// Service name
	Service string
	// Name running instance name
	Name string
	// Level logging level
	Level zerolog.Level
	// CallerLevel caller setting level
	CallerLevel zerolog.Level
	// Disable console color in debug mode
	DisableConsoleColor bool
	// Stdout log json in stdout
	Stdout bool
	// FileDir write log to dir
	FileDir string
}

// MakeSimpleZeroLogger create a new simple logger using zero logger
func MakeSimpleZeroLogger() *ZeroLogger {
	return &ZeroLogger{logger: log.Logger}
}

// MakeZeroLogger create a new logger using zero logger
func MakeZeroLogger(c LoggingConfig) (*ZeroLogger, error) {
	l := ZeroLogger{}
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.DisableSampling(true)
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.999Z07:00"
	zerolog.TimestampFunc = func() time.Time { return time.Now().UTC() }

	var err error
	logWriters := make([]io.Writer, 0, 3)
	if len(c.FileDir) > 0 {
		logName := c.Name
		if len(logName) == 0 {
			logName = c.Service
		}
		l.logFile, err = OpenLogFile(c.FileDir, logName)
		if err != nil {
			return nil, err
		}
		logWriters = append(logWriters, l.logFile)
	}

	zerolog.SetGlobalLevel(c.Level)

	if c.Stdout {
		logWriters = append(logWriters, os.Stdout)
	} else if c.Level == zerolog.DebugLevel {
		logWriters = append(logWriters, zerolog.ConsoleWriter{Out: os.Stderr, NoColor: c.DisableConsoleColor})
	}

	if len(logWriters) == 0 {
		return nil, errors.New("no log writer available")
	}

	logContext := zerolog.
		New(zerolog.MultiLevelWriter(logWriters...)).
		With().Timestamp().Str("service", c.Service)
	if len(c.Name) > 0 {
		logContext = logContext.Str("process", c.Name)
	}
	l.logger = logContext.Logger()
	l.callerLevel = c.CallerLevel

	return &l, nil
}

// CloseLogger close the logger
func (l *ZeroLogger) CloseLogger() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}
