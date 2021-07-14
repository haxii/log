package main

import (
	"fmt"
	"io"
	"time"

	"github.com/haxii/log"
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

	zeroLogger.Raw([]byte("this is a raw message, which should be logged in string format"), "")
	zeroLogger.Raw([]byte(`{"this is":{"a raw message":"which should be","logged":"in raw JSON format"}}`), "")
	if !zeroLogger.IsProduction() {
		zeroLogger.Debug("Example Client", "this is a %s", "debug output")
	}
	if !zeroLogger.IsProduction() {
		zeroLogger.Info("Example Client", "this is a %s", "debug output")
	}
	zeroLogger.Error(io.EOF, "this is a %s", "error output with EOF error")
	zeroLogger.Fatal(io.EOF, "this is a %s", "fatal output with EOF error")
}
