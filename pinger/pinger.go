// Copyright (c) 2018 codeliveroil. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

package pinger

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

// Pinger defines the construct that can be used to
// carry out a TCP or UDP based ping.
type Pinger struct {
	// Host is the host address to be pinged. This can be
	// in IP or name form.
	Host string

	// Port is the listening port on the host that we'd like
	// to ping.
	Port int

	// Protocol is the network protocol to be used. Valid values
	// are tcp and udp networks defined in net.Dial()
	// Known networks are "tcp", "tcp4" (IPv4-only), "tcp6" (IPv6-only),
	// "udp", "udp4" (IPv4-only), "udp6" (IPv6-only)
	Protocol string

	// Wait waits for a response from the server before considering the
	// ping to be successful. Since UDP is connectionless, set this to
	// true to ensure that the ping has actually established contact
	// with the host.
	Wait bool

	// PayloadSize is the number of bytes to send in each ping.
	PayloadSize int

	// Interval is the interval between pings.
	Interval time.Duration

	// TTL is the time to live for the packet in time format
	// as opposed to number of hops. If a ping does not complete
	// within the TTL, then the packet is assumed to be dropped.
	TTL time.Duration

	// MaxPings specifies the number of ping attempts to try
	// before returning.
	MaxPings int

	// DNSServer can be specified if a DNS server other than the
	// system default should be used. This is useful in OS'es that
	// don't typically have the standard DNS server configurations
	// (like /etc/resolv.conf) like on the Android OS.
	DNSServer string

	// Log write ping results and can be used to create a Linux
	// ping like application or can be used for debugging.
	Log func(string)

	interrupted bool
}

// Result contains the details of a ping operation.
type Result struct {
	Received int
	Dropped  int
}

// Ping pings the host and port defined in the Pinger. The Result is
// updated with each ping so it can be used at any time to compute the
// statistics so far. This comes in handy when Ctrl+C is trapped
// but you'd rather not wait for Ping() to return if it is stuck on
// a long TTL.
func (p *Pinger) Ping(r *Result) error {
	p.interrupted = false
	count := 0

	var payload []byte
	for i := 0; i < p.PayloadSize; i++ {
		payload = append(payload, 0x0a)
	}

	wrap := func(conn net.Conn, err error) {
		if conn != nil {
			conn.Close()
		}
		if err != nil {
			r.Dropped++
			p.Log(fmt.Sprintf("%v for seq=%d", err, count))
		}
		count++
		if count >= p.MaxPings {
			p.interrupted = true
		}
		time.Sleep(p.Interval)
	}

	if p.DNSServer != "" {
		dns := p.DNSServer
		if !strings.Contains(dns, ":") {
			dns += ":53"
		}
		_, err := net.DialTimeout(p.Protocol, dns, p.TTL)
		if err != nil {
			return errors.New("dns server not reachable: " + dns)
		}
		net.DefaultResolver = &net.Resolver{
			PreferGo: false,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return net.Dial(network, dns)
			},
		}
	}

	firstPacket := true
	for !p.interrupted {
		start1 := time.Now()
		conn, err := net.DialTimeout(p.Protocol, fmt.Sprintf("%s:%d", p.Host, p.Port), p.TTL)
		stop1 := time.Now()
		if err != nil {
			wrap(nil, err)
			continue
		}
		ip := conn.RemoteAddr().String()
		if firstPacket {
			firstPacket = false
			p.Log(fmt.Sprintf("PING %s (%s): %d data bytes", p.Host, ip, p.PayloadSize))
		}

		conn.SetDeadline(time.Now().Add(p.TTL))
		start2 := time.Now()
		n, err := conn.Write(payload)
		stop2 := time.Now()
		if err != nil {
			wrap(conn, err)
			continue
		}
		if n != p.PayloadSize {
			wrap(conn, errors.New(fmt.Sprintf("partial payload written (size=%d)", n)))
			continue
		}

		rtt := stop1.Sub(start1) + stop2.Sub(start2)
		payloadSize := p.PayloadSize
		payloadDirection := "to"
		if p.Wait {
			buf := make([]byte, 1024)
			start3 := time.Now()
			nr, err := conn.Read(buf)
			stop3 := time.Now()
			if err != nil {
				wrap(conn, err)
				continue
			}
			if nr <= 0 {
				wrap(conn, errors.New("no packet received"))
				continue
			}
			rtt += stop3.Sub(start3)
			payloadSize = nr
			payloadDirection = "from"
		}

		r.Received++
		p.Log(fmt.Sprintf("%d bytes %s %s %s_seq=%d time=%.3f ms",
			payloadSize, payloadDirection, ip, p.Protocol, count, float64(rtt)/float64(time.Millisecond)))
		wrap(conn, nil)
	}

	return nil
}

// Interrupt can be used to interrupt the Ping() operation.
// If the Pinger is in the middle of a ping, then it will
// need to time out first, for Ping() to return.
func (p *Pinger) Interrupt() {
	p.interrupted = true
}
