package log

import (
	"io"
	"os"

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
}

// GetZeroLogger returns the zero logger instance for advanced usage
func (l *ZeroLogger) GetZeroLogger() zerolog.Logger {
	return l.logger
}

const defaultCallSkip = 2

// Rawf implements raw logger interface
func (l *ZeroLogger) Rawf(rawMessage []byte, format string, v ...interface{}) {
	l.rawf(defaultCallSkip, rawMessage, format, v...)
}
func (l *ZeroLogger) rawf(callSkip int, rawMessage []byte, format string, v ...interface{}) {
	if json.Valid(rawMessage) {
		rawMessageInJSON := json.RawMessage(rawMessage)
		l.logger.WithLevel(zerolog.NoLevel).Caller(callSkip).Interface("raw", rawMessageInJSON).Msgf(format, v...)
	} else {
		l.logger.WithLevel(zerolog.NoLevel).Caller(callSkip).Bytes("raw", rawMessage).Msgf(format, v...)
	}
}

// Debugf implements debug logger interface
func (l *ZeroLogger) Debugf(format string, v ...interface{}) {
	l.debugf(defaultCallSkip, format, v...)
}
func (l *ZeroLogger) debugf(callSkip int, format string, v ...interface{}) {
	l.logger.Debug().Caller(callSkip).Msgf(format, v...)
}

// Infof implements info logger interface
func (l *ZeroLogger) Infof(format string, v ...interface{}) {
	l.infof(defaultCallSkip, format, v...)
}
func (l *ZeroLogger) infof(callSkip int, format string, v ...interface{}) {
	l.logger.Info().Caller(callSkip).Msgf(format, v...)
}

// Errorf implements error logger interface
func (l *ZeroLogger) Errorf(err error, format string, v ...interface{}) {
	l.errorf(defaultCallSkip, err, format, v...)
}
func (l *ZeroLogger) errorf(callSkip int, err error, format string, v ...interface{}) {
	l.logger.Error().Caller(callSkip).Stack().Err(err).Msgf(format, v...)
}

// Fatalf make a fatal return
func (l *ZeroLogger) Fatalf(err error, format string, v ...interface{}) {
	l.fatalf(defaultCallSkip, err, format, v...)
}
func (l *ZeroLogger) fatalf(callSkip int, err error, format string, v ...interface{}) {
	l.logger.Panic().Caller(callSkip).Stack().Err(err).Msgf(format, v...)
}

// LoggingConfig helper for a logging destination
type LoggingConfig struct {
	// Service service name
	Service string
	// Level logging level
	Level zerolog.Level
	// Disable console color
	DisableConsoleColor bool
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

	var err error
	logWriters := make([]io.Writer, 0, 3)
	if len(c.FileDir) > 0 {
		l.logFile, err = OpenLogFile(c.FileDir, c.Service)
		if err != nil {
			return nil, err
		}
		logWriters = append(logWriters, l.logFile)
	}

	zerolog.SetGlobalLevel(c.Level)

	if c.Level == zerolog.DebugLevel {
		logWriters = append(logWriters, zerolog.ConsoleWriter{Out: os.Stderr, NoColor: c.DisableConsoleColor})
	}

	if len(logWriters) == 0 {
		return nil, errors.New("no log writer available")
	}

	l.logger = zerolog.
		New(zerolog.MultiLevelWriter(logWriters...)).
		With().Timestamp().Str("service", c.Service).Logger()

	return &l, nil
}

// CloseLogger close the logger
func (l *ZeroLogger) CloseLogger() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}
