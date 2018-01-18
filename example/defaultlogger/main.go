package main

import (
	"io"

	"github.com/haxii/log"
)

func main() {
	defaultLoggerExample()
}

func defaultLoggerExample() {
	defaultLogger := &log.DefaultLogger{}
	defaultLogger.Debug("Example Client", "this is a %s", "debug output")
	defaultLogger.Info("Example Client", "this is a %s", "debug output")
	defaultLogger.Error("Example Client", io.EOF, "this is a %s", "error output with EOF error")
	defaultLogger.Fatal("Example Client", io.EOF, "this is a %s", "fatal output with EOF error")
}
