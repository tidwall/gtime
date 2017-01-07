package gtime

import (
	"testing"
	"time"
)

func TestNow(t *testing.T) {
	Sync(time.Second)
	t1 := Now()
	t2 := Now()
	if t1.After(t2) {
		t.Fatalf("time out of order, %v > %v", t1, t2)
	}
}
