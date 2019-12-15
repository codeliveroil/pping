// Copyright (c) 2018 codeliveroil. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/codeliveroil/niceflags"
	"github.com/codeliveroil/pping/pinger"
)

const version = "1.1"

func main() {
	args := os.Args
	flags := niceflags.NewFlags(
		os.Args[0],
		"pping - Protocol Ping",
		"Tool to simulate TCP and UDP pings. This can also be used as a port scanner.",
		"[options] host port",
		"help",
		false)
	flags.Examples = []string{
		"-s 128 google.com 80",
		"-p udp -w -c 5 -t 1000 myserver.com 8085",
	}

	payloadSize := flags.Int("s", 64, "Payload `size` in bytes.")
	interval := flags.Int("i", 1000, "Interval `time` between pings in ms.")
	ttl := flags.Int("t", 10000, "Max `time`-to-live for each ping (in ms) before moving on to the next attempt.")
	protocol := flags.String("p", "tcp", "Specify `protocol` to use. Valid values are tcp and udp and their 4 and 6 "+
		"counterparts (e.g. tcp6 which implies tcp6-only). Note that since UDP is connectionless, pings will always "+
		"succeed even if there is no server listening on the given host and port. Use -w to wait for a response.")
	wait := flags.Bool("w", false, "Wait for a response from the server. Ideally, this should be set when the protocol is set to udp.")
	max := flags.Int("c", math.MaxInt32, "Stop after sending specified `num`ber of pings.")
	dns := flags.String("d", "", "DNS `server` IP address to use. This can be specified for name resolution on systems that don't use "+
		"the traditional DNS server configurations such as /etc/resolv.conf.")
	version := flags.Bool("v", false, "Display version.")

	check(flags.Parse(args[1:]))
	flags.Help()

	argc := len(args)
	if *version {
		fmt.Println(version)
		os.Exit(0)
	}

	if argc == 1 {
		flags.Usage()
		os.Exit(0)
	}

	if argc < 3 {
		niceflags.PrintErr("invalid usage\n")
		flags.Usage()
		os.Exit(1)
	}

	host := args[argc-2]
	var port, err = strconv.Atoi(args[argc-1])
	check(err)

	if host == "" {
		niceflags.PrintErr("host not specified\n")
		flags.Usage()
		os.Exit(1)
	}
	if port <= 0 {
		niceflags.PrintErr("invalid port specified\n")
		flags.Usage()
		os.Exit(1)
	}

	p := pinger.Pinger{
		Host:        host,
		Port:        port,
		Protocol:    *protocol,
		Wait:        *wait,
		PayloadSize: *payloadSize,
		Interval:    time.Duration(*interval) * time.Millisecond,
		TTL:         time.Duration(*ttl) * time.Millisecond,
		MaxPings:    *max,
		DNSServer:   *dns,
		Log:         func(msg string) { fmt.Println(msg) },
	}

	r := &pinger.Result{}
	finish := func() {
		fmt.Println()
		fmt.Printf("--- %s:%d ping statistics ---\n", p.Host, p.Port)
		total := r.Received + r.Dropped
		pktLoss := float64(r.Dropped) / float64(total) * 100
		fmt.Printf("%d packets transmitted, %d packets received, %.2f%% packet loss\n", total, r.Received, pktLoss)
		code := 0
		if r.Dropped > 0 {
			code = 2
			if r.Received == 0 {
				code++
			}
		}
		os.Exit(code)
	}

	ic := make(chan os.Signal, 1)
	signal.Notify(ic, os.Interrupt)
	go func() {
		for sig := range ic {
			_ = sig
			p.Interrupt()
			finish()
		}
	}()

	err = p.Ping(r)
	check(err)
	finish()
}

// check handles errors.
func check(err error) {
	if err != nil {
		niceflags.PrintErr("%v\n", err)
		os.Exit(1)
	}
}
