package main

import (
	"github.com/haxii/log"
	"io"
)

func main() {
	defaultLoggerExample()
}

func defaultLoggerExample() {
	defaultLogger := &log.DefaultLogger{}
	defaultLogger.Rawf([]byte("this is a raw message, which should be logged in plain format"), "")
	defaultLogger.Rawf([]byte(`{"this is":{"a raw message":"which should be","logged":"in plain format"}}`), "")
	defaultLogger.Debugf("this is a %s", "debug output")
	defaultLogger.Infof("this is a %s", "debug output")
	defaultLogger.Errorf(io.EOF, "this is a %s", "error output with EOF error")
	defaultLogger.Fatalf(io.EOF, "this is a %s", "fatal output with EOF error")
}
