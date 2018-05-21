package log

import (
	"errors"
	"math"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

/* implements TCP and UDP input for log */

type logstashWriter struct {
	conn net.Conn

	inputType LogstashInputType

	maxRetries    int
	retryInterval time.Duration

	lock   sync.RWMutex
	status int32
}

// status enum
const (
	statusOnline int32 = iota
	statusOffline
	statusReconnecting
)

// LogstashInputType log stash input type tcp or udp
type LogstashInputType int

const (
	// LogstashInputTypeTCP tcp input
	LogstashInputTypeTCP LogstashInputType = iota
	// LogstashInputTypeUDP udp input
	LogstashInputTypeUDP
)

// LogstashConfig config for logstash
type LogstashConfig struct {
	Type LogstashInputType
	Addr string
}

var errUnknownInputType = errors.New("unsupported logstash input type")

func makeLogstashWriter(c LogstashConfig) (*logstashWriter, error) {
	var conn net.Conn
	switch c.Type {
	case LogstashInputTypeTCP:
		addr, err := net.ResolveTCPAddr("tcp", c.Addr)
		if err != nil {
			return nil, err
		}
		conn, err = net.DialTCP("tcp", nil, addr)
		if err != nil {
			return nil, err
		}
	case LogstashInputTypeUDP:
		addr, err := net.ResolveUDPAddr("udp", c.Addr)
		if err != nil {
			return nil, err
		}
		conn, err = net.DialUDP("udp", udpSrcAddr, addr)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errUnknownInputType
	}

	return &logstashWriter{
		conn:          conn,
		inputType:     c.Type,
		maxRetries:    10,
		retryInterval: 10 * time.Millisecond,
		lock:          sync.RWMutex{},
	}, nil
}

func (l *logstashWriter) reconnect() error {
	// set the shared status to 'reconnecting', if it's already the case, return early,
	// something's already trying to reconnect
	if !atomic.CompareAndSwapInt32(&l.status, statusOffline, statusReconnecting) {
		return nil
	}

	serverAddr := l.conn.RemoteAddr()
	conn, err := net.DialTCP(serverAddr.Network(), nil, serverAddr.(*net.TCPAddr))
	if err != nil {
		// reset shared status to offline
		defer atomic.StoreInt32(&l.status, statusOffline)
		return err
	}

	// set new TCP socket
	l.conn.Close()
	l.conn = conn

	// we're back online, set shared status accordingly
	atomic.StoreInt32(&l.status, statusOnline)

	return nil
}

var udpSrcAddr = &net.UDPAddr{IP: net.IPv4zero, Port: 0}

// Write implements io.Write interface using logstash TCP & UDP input
func (l *logstashWriter) Write(p []byte) (n int, err error) {
	if l.inputType == LogstashInputTypeUDP {
		return l.conn.Write(p)
	}

	defer func() {
		recover()
	}()

	l.lock.RLock()
	defer l.lock.RUnlock()

	for i := 0; i < l.maxRetries; i++ {
		if atomic.LoadInt32(&l.status) == statusOnline {
			n, err := l.conn.Write(p)
			if err == nil {
				return n, err
			}
			switch e := err.(type) {
			case *net.OpError:
				if realErrNo, ok := e.Err.(syscall.Errno); ok {
					if realErrNo == syscall.ECONNRESET ||
						realErrNo == syscall.EPIPE {
						atomic.StoreInt32(&l.status, statusOffline)
					}
				} else if realErr, ok := e.Err.(*os.SyscallError); ok {
					if realErr.Err == syscall.ECONNRESET ||
						realErr.Err == syscall.EPIPE {
						atomic.StoreInt32(&l.status, statusOffline)
					}
				} else {
					return n, err
				}
			default:
				if err.Error() == "EOF" {
					atomic.StoreInt32(&l.status, statusOffline)
				} else {
					return n, err
				}
			}
		} else if atomic.LoadInt32(&l.status) == statusOffline {
			if err := l.reconnect(); err != nil {
				return -1, err
			}
		}

		// exponential backoff
		if i < (l.maxRetries - 1) {
			time.Sleep(l.retryInterval * time.Duration(math.Pow(2, float64(i))))
		}
	}

	return -1, ErrMaxConnRetries
}

// ErrMaxConnRetries max connection retries exceeded
var ErrMaxConnRetries = errors.New("max connection retries exceeded")
