package log

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
)

// ZeroLogger implemented logger using zerolog
type ZeroLogger struct {
	isProduction bool
	logger       zerolog.Logger
	logFile      *os.File
	logstash     *logstashWriter
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
		l.logFile, err = l.openLogFile(c.FileDir, service)
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
		l.logstash, err = makeLogstashWriter(*c.Logstash)
		if err != nil {
			return nil, err
		}
		if c.LazyLogging != nil {
			diodeLogstash := diode.NewWriter(l.logstash,
				c.LazyLogging.DiodeSize, c.LazyLogging.PoolInterval, nil)
			logWriters = append(logWriters, diodeLogstash)
		} else {
			logWriters = append(logWriters, l.logstash)
		}
	}

	if debug {
		logWriters = append(logWriters, zerolog.ConsoleWriter{Out: os.Stderr})
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}

	if len(logWriters) == 0 {
		return nil, errors.New("no log writer avaliable")
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

func (l *ZeroLogger) openLogFile(logdir, serviceName string) (*os.File, error) {
	timeNOW := func() string {
		return time.Now().Format("2006-01-02-15.04.05.999999999")
	}

	logFileName := serviceName + ".log"
	logFilePath := filepath.Join(logdir, logFileName)

	if l.fileExists(logFilePath) {
		os.Rename(logFilePath, filepath.Join(logdir, serviceName+"."+timeNOW()+".log"))
	}
	logWriter, err := l.makefile(logdir, logFileName)
	if err != nil {
		return nil, err
	}
	return logWriter, nil
}

// log file helper
func (l *ZeroLogger) makefile(dir string, filename string) (f *os.File, e error) {
	if err := l.createDirectories(dir); err != nil {
		return nil, err
	}
	filePath := filepath.Join(dir, filename)
	fileWriter, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return fileWriter, nil
}

func (l *ZeroLogger) createDirectories(dir string) error {
	if fi, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("%v (checking directory)", err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("%v (creating directory)", err)
		}
	} else if !fi.IsDir() {
		return fmt.Errorf("destination path is not directory")
	}
	return nil
}

func (l *ZeroLogger) fileExists(path string) bool {
	if len(path) == 0 {
		return false
	}

	if fi, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else if fi.IsDir() {
		return false
	}
	return true
}
