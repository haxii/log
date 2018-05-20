package log

import (
	"errors"
	"net"
	"sync"
	"syscall"
	"time"
)

/* implements TCP and UDP input for log */

type logstashWriter struct {
	conn net.Conn

	inputType LogstashInputType

	maxRetries    int
	retryInterval time.Duration

	lock sync.RWMutex
}

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
	l.lock.Lock()
	defer l.lock.Unlock()

	serverAddr := l.conn.RemoteAddr()
	var conn net.Conn
	var err error
	switch l.inputType {
	case LogstashInputTypeTCP:
		conn, err = net.DialTCP(serverAddr.Network(), nil, serverAddr.(*net.TCPAddr))
	case LogstashInputTypeUDP:
		conn, err = net.DialUDP(serverAddr.Network(), udpSrcAddr, serverAddr.(*net.UDPAddr))
	default:
		err = errUnknownInputType
	}
	if err != nil {
		return err
	}

	l.conn.Close()
	l.conn = conn
	return nil
}

var udpSrcAddr = &net.UDPAddr{IP: net.IPv4zero, Port: 0}

// Write implements io.Write interface using logstash UDP input
func (l *logstashWriter) Write(p []byte) (n int, err error) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	disconnected := false

	t := l.retryInterval
	for i := 0; i < l.maxRetries; i++ {
		if disconnected {
			time.Sleep(t)
			t *= 2
			l.lock.RUnlock()
			if err := l.reconnect(); err != nil {
				switch e := err.(type) {
				case *net.OpError:
					if e.Err.(syscall.Errno) == syscall.ECONNREFUSED {
						disconnected = true
						l.lock.RLock()
						continue
					}
					return -1, err
				default:
					return -1, err
				}
			} else {
				disconnected = false
			}
			l.lock.RLock()
		}
		n, err := l.conn.Write(p)
		if err == nil {
			return n, err
		}
		switch e := err.(type) {
		case *net.OpError:
			if e.Err.(syscall.Errno) == syscall.ECONNRESET ||
				e.Err.(syscall.Errno) == syscall.EPIPE {
				disconnected = true
			} else {
				return n, err
			}
		default:
			if err.Error() == "EOF" {
				disconnected = true
			} else {
				return n, err
			}
		}
		t *= 2
	}

	return -1, ErrMaxConnRetries
}

// ErrMaxConnRetries max connection retries exceeded
var ErrMaxConnRetries = errors.New("max connection retries exceeded")
