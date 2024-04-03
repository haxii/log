package recover

import (
	"github.com/haxii/log/v2"
	"github.com/haxii/log/v2/recover/internal"
	"github.com/pkg/errors"
	"net"
	"net/http"
	"os"
	"strings"
)

func Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer Func(func(connBroken bool, recoveredErr error) {
			if !connBroken {
				http.Error(w, recoveredErr.Error(), http.StatusInternalServerError)
			}
		})
		h.ServeHTTP(w, req)
	})
}

func Func(h func(broken bool, recoveredErr error)) {
	if err := recover(); err != nil {
		// Check for a broken connection, as it is not really a
		// condition that warrants a panic stack trace.
		var brokenPipe bool
		if ne, ok := err.(*net.OpError); ok {
			if se, ok := ne.Err.(*os.SyscallError); ok {
				if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
					strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
					brokenPipe = true
				}
			}
		}
		_stack := internal.Stack(3)
		e := errors.Errorf("%s", err)
		if brokenPipe {
			log.Errorf(e, "broken pipe occurred")
		} else {
			log.Errorf(e, "serv panic recovered:\n%s", _stack)
		}
		if h != nil {
			h(brokenPipe, e)
		}
	}
}
