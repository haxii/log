package log

import (
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAutoReconnect(t *testing.T) {
	// open a server socket
	s, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	// save the original port
	addr := s.Addr()
	// connect a client to the server
	c, err := makeLogstashWriter(
		LogstashConfig{
			Type: LogstashInputTypeTCP,
			Addr: addr.String(),
			KeepAliveCheckInterval: time.Second,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("server & client started")
	if _, err := c.Write([]byte("hello, world!\n")); err != nil {
		t.Fatal(err)
	}

	if status := atomic.LoadInt32(&c.status); status != statusOnline {
		t.Fatalf("unexpected status %+v, expecting online", status)
	}
	t.Logf("connection status online")

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Second)
	t.Log("server stopped 3 secs ago")

	if status := atomic.LoadInt32(&c.status); status != statusOffline {
		t.Fatalf("unexpected status %+v, expecting offline", status)
	}
	t.Logf("connection status offline")

	t.Logf("invoke reconnection")
	go func() {
		c.Write([]byte("hello, world!\n"))
	}()

	s, err = net.Listen("tcp", addr.String())
	time.Sleep(10 * time.Second)
	t.Log("server started 10 secs ago")

	if status := atomic.LoadInt32(&c.status); status != statusOnline {
		t.Fatalf("unexpected status %+v, expecting online", status)
	}
	t.Logf("connection status online")

	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Second)
}

func TestMultipleWrite(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// open a server socket
	s, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	// save the original port
	addr := s.Addr()

	// connect a client to the server
	c, err := makeLogstashWriter(
		LogstashConfig{
			Type: LogstashInputTypeTCP,
			Addr: addr.String(),
			KeepAliveCheckInterval: time.Second,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	//defer c.Close()

	// shut down and boot up the server randomly
	var swg sync.WaitGroup
	swg.Add(1)
	go func() {
		defer swg.Done()
		for i := 0; i < 5; i++ {
			t.Log("server up")
			time.Sleep(time.Millisecond * 100 * time.Duration(rand.Intn(20)))
			if err := s.Close(); err != nil {
				t.Fatal(err)
			}
			t.Log("server down")
			time.Sleep(time.Millisecond * 100 * time.Duration(rand.Intn(20)))
			s, err = net.Listen("tcp", addr.String())
			if err != nil {
				t.Fatal(err)
			}
		}
	}()

	// client writes to the server and reconnects when it has to
	// this is the interesting part
	var cwg sync.WaitGroup
	cwg.Add(1)
	go func() {
		defer cwg.Done()
		for {
			if _, err := c.Write([]byte("hello, world!\n")); err != nil {
				if err == ErrMaxConnRetries {
					t.Log("client leaving, reached retry limit")
					return
				}
				t.Fatal(err)
			}
			t.Log("client says hello!")
		}
	}()

	// terminates the server indefinitely
	swg.Wait()
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	// wait for the client to give up
	cwg.Wait()
}
