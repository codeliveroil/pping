# pping - TCP and UDP Pinger

pping (Protocol Ping) is a command line utility that can simulate ICMP-like pings for the TCP and UDP protocols.

<img src="resources/readme/screenshot.png" alt="Screenshot" width="65%" height="65%"/>

<img src="resources/readme/demo.gif" alt="GIF demo" width="65%" height="65%"/>

## Installation

#### macOS
```
brew tap codeliveroil/apps
brew install pping
```

#### Other
Download the [latest release](../../releases/latest) for your operating system and machine architecture. If one is not available, you can easily [compile from source](#compile-from-source).

## Usage
```
pping -help
```

#### Examples
```
pping myserver.com 55004
pping -p udp -w 192.168.1.9 40001
pping -d 8.8.8.8 example.com 8080
```

## Library API for GO

```golang
import "github.com/codeliveroil/pping/pinger"

...

p := pinger.Pinger{
	Host:        "google.com",
	Port:        80,
	Protocol:    "tcp",
	Wait:        false,
	PayloadSize: 64,
	Interval:    1 * time.Second,
	TTL:         10 * time.Second,
	MaxPings:    5,
	DNSServer:   "",
	Log:         func(msg string) { fmt.Println(msg) },
}

res := &pinger.Result{}
err := p.Ping(res)
if err != nil {
	// handle error
}

fmt.Printf("Received=%d, Dropped=%d, Total=%d\n", res.Received, res.Dropped, res.Received+res.Dropped)
```

## Compile from source

### Setup
1. Install [Go](https://golang.org/)
1. Clone this repository

### Build for your current platform
```
make
make install
```

### Cross compile for a different platform
1. Build
	```
	make platform
	```
1. Follow the prompts and select a platform
1. The binary will be available in the `build` folder
