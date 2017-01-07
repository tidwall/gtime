// package gtime allows for syncing with Google time. This is useful for
// applications that run on servers that have a high risk for time drift,
// such as containers, virtual servers, and cloud providers.
package gtime

import (
	"io"
	"net"
	"strings"
	"sync"
	"time"
	_ "unsafe"
)

//go:linkname nanotime runtime.nanotime
func nanotime() time.Duration

var (
	gmu   sync.RWMutex
	gnano time.Duration
	gtime time.Time
)

// Sync will sync the time with Google servers. If the operation was successful
// then every following Now() call will return Google time.
// Returns an error if time cannot be fetched or the timeout has been reached.
func Sync(timeout time.Duration) error {
	t, nano, err := getNow(timeout)
	if err != nil {
		return err
	}
	gmu.Lock()
	gtime, gnano = t, nano
	gmu.Unlock()
	return nil
}

// MustSync will attempt to sync with Google servers. It will try over and over
// again until the timeout has been reached. It will panic if the timeout is
// reached. If the operation was successful then every following Now() call
// will return Google time.
func MustSync(timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for {
		timeout := deadline.Sub(time.Now())
		if err := Sync(timeout); err != nil {
			if deadline.Sub(time.Now()) < 0 {
				panic(err)
			}
			time.Sleep(time.Millisecond * 50)
			continue
		}
		break
	}
}

// Now returns the current Google time.
// Local system time is returned if Sync or MustSync has not been
// succesfully called.
func Now() time.Time {
	gmu.RLock()
	t, nano := gtime, gnano
	gmu.RUnlock()
	if nano == 0 {
		panic("time has not been synced")
	}
	return t.Add(time.Duration(nanotime() - nano))
}

func getNow(timeout time.Duration) (
	t time.Time, nano time.Duration, err error,
) {
	deadline := time.Now().Add(timeout)
	// connect to public google.com on port 80. This should resolve globally
	// keeping the hops down regardless of where in the world we are.
	c, err := net.DialTimeout("tcp", "google.com:80", timeout)
	if err != nil {
		return time.Time{}, 0, err
	}
	defer c.Close()
	err = c.SetWriteDeadline(deadline)
	if err != nil {
		return time.Time{}, 0, err
	}
	// Using a dash a the resource path with a head ensures that a 404 is
	// returned very quickly, which is what we want. It's likely that the
	// request will fail at the proxy level instead of making it to an
	// application server.
	_, err = io.WriteString(c, "HEAD - HTTP/1.0\r\n\r\n")
	if err != nil {
		return time.Time{}, 0, err
	}
	b := make([]byte, 128)
	err = c.SetReadDeadline(deadline)
	if err != nil {
		return time.Time{}, 0, err
	}
	n, err := c.Read(b)
	if err != nil {
		return time.Time{}, 0, err
	}
	// get out server clock prior to parsing the response. This value will
	// be used as the seed to sync against for all following Now calls.
	nano = nanotime()
	var dts string
	for _, line := range strings.Split(string(b[:n]), "\r\n") {
		if strings.HasPrefix(line, "Date:") {
			dts = strings.TrimSpace(line[5:])
			break
		}
	}
	t, err = time.Parse(time.RFC1123, dts)
	if err != nil {
		return time.Time{}, 0, err
	}
	return t.Local(), nano, nil
}
