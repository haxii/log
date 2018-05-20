package main

import (
	"io"

	"github.com/haxii/log"
)

func main() {
	zeroLoggerExample()
}

func zeroLoggerExample() {
	zeroLogger, err := log.MakeZeroLogger(true,
		log.LoggingConfig{
			FileDir: "/tmp/log",
			Logstash: &log.LogstashConfig{
				Type: log.LogstashInputTypeTCP,
				Addr: "10.1.3.8:8084",
			},
		},
		"ExampleService")
	if err != nil {
		panic(err)
	}
	defer zeroLogger.CloseLogger()
	zeroLogger.Debug("Example Client", "this is a %s", "debug output")
	zeroLogger.Info("Example Client", "this is a %s", "debug output")
	zeroLogger.Error("Example Client", io.EOF, "this is a %s", "error output with EOF error")
	zeroLogger.Fatal("Example Client", io.EOF, "this is a %s", "fatal output with EOF error")
}
