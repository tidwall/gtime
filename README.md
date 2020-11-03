**This project has been archived. Please check out [tidwall/rtime](https://github.com/tidwall/rtime) for a better internet time library.**

gtime
=====

[![GoDoc](https://img.shields.io/badge/api-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/tidwall/gtime)

A Go package that helps to keep your application time in sync with Google server time.

This can be very useful for long running applications which reside on systems
that may have a higher risk for clock drift. Such as containers, virtual servers,
and cloud providers.

Getting
-------

```
go get github.com/tidwall/gtime
```

API
---

```go
// Now returns the current Google time. 
// Local system time is returned until Sync or MustSync has been succesfully called.
gtime.Now() time.Time

// Sync will sync the time with Google servers. 
gtime.Sync(timeout time.Duration) error

// MustSync will attempt to sync with Google servers. 
// This operation will try over and over again until the time has successfully 
// synced or the timeout has been reached. A timeout will panic.
gtime.MustSync(timeout time.Duration)
```

Example
-------

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/tidwall/gtime"
)

func main() {
	// Require a sync with Google time. This will panic if it does not
	// succeed in one minute. Optionally you can just call gtime.Sync() and
	// handle the error.
	gtime.MustSync(time.Minute)

	// Spin up a background routine to keep the system in sync. This example 
	// syncs with Google every 5 minutes.
	go func() {
		for range time.NewTicker(time.Minute * 5).C {
			gtime.Sync(time.Minute)
		}
	}()

	// Create a little HTTP server that responds to all requests with
	// Google and Local time.
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "Google: %v\n Local: %v\n", gtime.Now(), time.Now())
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
```

Then from a terminal:

```bash
$ curl http://localhost:8080
Google: 2017-01-07 15:45:02.288466989 -0700 MST
 Local: 2017-01-07 15:45:02.529567671 -0700 MST
```

Contact
-------
Josh Baker [@tidwall](http://twitter.com/tidwall)

License
-------
Gtime source code is available under the MIT [License](/LICENSE).
