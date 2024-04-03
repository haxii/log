package recover

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestRecover(t *testing.T) {
	http.ListenAndServe(":8123", Handler(http.HandlerFunc(func(
		w http.ResponseWriter, req *http.Request) {
		x, _ := strconv.ParseInt(req.Header.Get("x-test"), 10, 64)
		fmt.Fprintf(w, "%d", 100/x)
	})))
	// curl 127.0.0.1:8123
	// got error [ runtime error: integer divide by zero ]
	time.Sleep(time.Minute)
}
