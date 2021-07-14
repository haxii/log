package main

import (
	"fmt"
	"time"

	"github.com/haxii/log/v2"
	"github.com/pkg/errors"
)

func main() {
	zeroLoggerExample()
}

func zeroLoggerExample() {
	zeroLogger, err := log.MakeZeroLogger(true,
		log.LoggingConfig{
			FileDir: "/tmp/log",
			LazyLogging: &log.LazyLogging{
				DiodeSize:    1000,
				PoolInterval: 10 * time.Millisecond,
			},
		},
		"ExampleService")
	if err != nil {
		panic(err)
	}
	defer zeroLogger.CloseLogger()

	defer func() {
		if err := recover(); err != nil {
			// enough time for the logger to flush
			time.Sleep(11 * time.Millisecond)
			fmt.Println(err)
		}
	}()

	zeroLogger.Rawf([]byte("this is a raw message, which should be logged in string format"), "")
	zeroLogger.Rawf([]byte(`{"this is":{"a raw message":"which should be","logged":"in raw JSON format"}}`), "")
	if !zeroLogger.IsProduction() {
		zeroLogger.Debugf("this is a %s", "debug output")
	}
	if !zeroLogger.IsProduction() {
		zeroLogger.Infof("this is a %s", "debug output")
	}
	zeroLogger.Errorf(outer(), "this is a %s", "error output with EOF error")
	zeroLogger.Fatalf(outer(), "this is a %s", "fatal output with EOF error")
}

func inner() error {
	return errors.New("seems we have an error here")
}

func middle() error {
	err := inner()
	if err != nil {
		return err
	}
	return nil
}

func outer() error {
	err := middle()
	if err != nil {
		return err
	}
	return nil
}
