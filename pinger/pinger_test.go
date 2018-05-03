// Copyright (c) 2018 codeliveroil. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

package pinger

import (
	"fmt"
	"math"
	"net"
	"testing"
	"time"
)

type server struct {
	tcpl    net.Listener
	udpl    *net.UDPConn
	ready   chan int
	running bool
}

func newServer(t *testing.T, protocol string, port int) *server {

	s := &server{
		ready:   make(chan int, 1),
		running: true,
	}

	var err error
	switch protocol {
	case "tcp":
		s.tcpl, err = net.Listen(protocol, fmt.Sprintf("localhost:%d", port))
	case "udp":
		addr, err := net.ResolveUDPAddr(protocol, fmt.Sprintf("localhost:%d", port))
		checkErr(t, "Cannot resolve UDP address", err)
		s.udpl, err = net.ListenUDP(protocol, addr)
	}

	checkErr(t, "cannot start server", err)

	go func() {
		s.ready <- 0
		for s.running {
			switch protocol {
			case "tcp":
				conn, err := s.tcpl.Accept() //ignore error because we'll be interrupting the connection
				if err != nil && conn != nil {
					buf := make([]byte, 1024)
					n, err := conn.Read(buf) //ignore error here as well
					if err != nil {
						fmt.Println("Received: ", n)
						conn.Write([]byte("Message received."))
					}
					conn.Close()
				}
			case "udp":
				buf := make([]byte, 1024)
				if s.udpl != nil {
					n, addr, _ := s.udpl.ReadFromUDP(buf) //ignore error because we'll be interrupting the connection
					fmt.Println("Received: ", n)
					s.udpl.WriteTo([]byte("Message received."), addr) //ignore error here as well
				}
			}
		}
	}()

	return s
}

func (s *server) stop() {
	s.running = false
	if s.tcpl != nil {
		s.tcpl.Close()
	}
	if s.udpl != nil {
		s.udpl.Close()
	}
}

func TestPinger(t *testing.T) {
	testProtocol(t, "tcp", 50053)
	testProtocol(t, "udp", 50055)
}

func testProtocol(t *testing.T, protocol string, port int) {
	var err error

	//Start server
	s := newServer(t, protocol, port)
	<-s.ready

	pinger := Pinger{
		Host:        "127.0.0.1",
		Port:        port,
		Protocol:    protocol,
		Wait:        protocol == "udp",
		PayloadSize: 64,
		Interval:    20 * time.Millisecond,
		TTL:         200 * time.Millisecond,
		MaxPings:    5,
		DNSServer:   "",
		Log:         func(msg string) { fmt.Println(msg) },
	}

	// Test 0% packet loss
	var res Result
	err = pinger.Ping(&res)
	checkErr(t, "could not start ping client", err)
	compare(t, 5, res.Received)
	compare(t, 0, res.Dropped)

	// Test some packet loss
	res = Result{}
	pinger.MaxPings = math.MaxInt32
	pingerDone := make(chan int, 1)
	go func() {
		err = pinger.Ping(&res)
		checkErr(t, "could not start ping client", err)
		pingerDone <- 1
	}()
	time.Sleep(300 * time.Millisecond)
	s.stop()
	time.Sleep(300 * time.Millisecond)
	pinger.Interrupt()
	compare(t, true, res.Received > 0)
	compare(t, true, res.Dropped > 0)
	<-pingerDone

	// Test 100% packet loss
	res = Result{}
	pinger.MaxPings = 5
	pinger.TTL = 30 * time.Millisecond
	err = pinger.Ping(&res)
	checkErr(t, "could not start ping client", err)
	compare(t, 0, res.Received)
	compare(t, 5, res.Dropped)
}

func compare(t *testing.T, exp, got interface{}) {
	if exp != got {
		t.Errorf("expected: %v, got: %v", exp, got)
	}
}

func checkErr(t *testing.T, msg string, err error) {
	if err != nil {
		t.Error(msg+": ", err)
	}
}
