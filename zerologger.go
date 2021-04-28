package log

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
)

// ZeroLogger implemented logger using zerolog
type ZeroLogger struct {
	isProduction bool
	logger       zerolog.Logger
	logFile      *os.File
	logStash     *logstashWriter
}

// GetZeroLogger returns the zero logger instance for advanced usage
func (l *ZeroLogger) GetZeroLogger() zerolog.Logger {
	return l.logger
}

// IsProduction implements raw logger interface to indicate
// the production level, avoids the meaningless calculation in debug and info
func (l *ZeroLogger) IsProduction() bool {
	return l.isProduction
}

// Raw implements raw logger interface
func (l *ZeroLogger) Raw(rawMessage []byte, format string, v ...interface{}) {
	if json.Valid(rawMessage) {
		rawMessageInJSON := json.RawMessage(rawMessage)
		l.logger.WithLevel(zerolog.NoLevel).Interface("raw", rawMessageInJSON).Msgf(format, v...)
	} else {
		l.logger.WithLevel(zerolog.NoLevel).Bytes("raw", rawMessage).Msgf(format, v...)
	}
}

// Debug implements debug logger interface
func (l *ZeroLogger) Debug(who, format string, v ...interface{}) {
	l.logger.Debug().Str("who", who).Msgf(format, v...)
}

// Info implements info logger interface
func (l *ZeroLogger) Info(who, format string, v ...interface{}) {
	l.logger.Info().Str("who", who).Msgf(format, v...)
}

// Error implements error logger interface
func (l *ZeroLogger) Error(who string, err error, format string, v ...interface{}) {
	l.logger.Error().Err(err).Str("who", who).Msgf(format, v...)
}

// Fatal make a fatal return
func (l *ZeroLogger) Fatal(who string, err error, format string, v ...interface{}) {
	l.logger.Panic().Err(err).Str("who", who).Msgf(format, v...)
}

// LazyLogging lazy logging settings
type LazyLogging struct {
	// DiodeSize ring buffer size
	DiodeSize int
	// PoolInterval the interval Diode query for data
	PoolInterval time.Duration
}

// LoggingConfig helper for a logging destination
type LoggingConfig struct {
	// Disable console color
	DisableConsoleColor bool
	// FileDir write log to dir
	FileDir string
	// Logstash config
	Logstash *LogstashConfig
	// LazyLogging settings
	LazyLogging *LazyLogging
}

// ErrInvalidZeroLogConfig provided logging config is invalid
var ErrInvalidZeroLogConfig = errors.New("invalid logger config")

// MakeZeroLogger create a new logger using zero logger
func MakeZeroLogger(debug bool, c LoggingConfig, service string) (*ZeroLogger, error) {
	if c.LazyLogging != nil {
		if c.LazyLogging.DiodeSize <= 0 {
			return nil, ErrInvalidZeroLogConfig
		}
	}
	l := ZeroLogger{}
	l.isProduction = !debug
	zerolog.DisableSampling(true)
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.999Z07:00"

	var err error
	logWriters := make([]io.Writer, 0, 3)
	if len(c.FileDir) > 0 {
		l.logFile, err = OpenLogFile(c.FileDir, service)
		if err != nil {
			return nil, err
		}
		if c.LazyLogging != nil {
			diodeLogFile := diode.NewWriter(l.logFile,
				c.LazyLogging.DiodeSize, c.LazyLogging.PoolInterval, nil)
			logWriters = append(logWriters, diodeLogFile)
		} else {
			logWriters = append(logWriters, l.logFile)
		}
	}

	if c.Logstash != nil {
		l.logStash, err = makeLogstashWriter(*c.Logstash)
		if err != nil {
			return nil, err
		}
		if c.LazyLogging != nil {
			diodeLogstash := diode.NewWriter(l.logStash,
				c.LazyLogging.DiodeSize, c.LazyLogging.PoolInterval, nil)
			logWriters = append(logWriters, diodeLogstash)
		} else {
			logWriters = append(logWriters, l.logStash)
		}
	}

	if debug {
		logWriters = append(logWriters, zerolog.ConsoleWriter{Out: os.Stderr, NoColor: c.DisableConsoleColor})
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}

	if len(logWriters) == 0 {
		return nil, errors.New("no log writer available")
	}

	l.logger = zerolog.
		New(zerolog.MultiLevelWriter(logWriters...)).
		With().Timestamp().Str("service", service).Logger()

	return &l, nil
}

// CloseLogger close the logger
func (l *ZeroLogger) CloseLogger() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}
