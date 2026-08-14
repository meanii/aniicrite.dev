package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterBlocksThenExpires(t *testing.T) {
	now := time.Unix(0, 0)
	l := New(3, time.Minute)
	l.nowFn = func() time.Time { return now }

	// Under the limit: allowed after each failure until max reached.
	for i := range 3 {
		if !l.Allow("ip") {
			t.Fatalf("attempt %d should be allowed", i)
		}
		l.Fail("ip")
	}
	if l.Allow("ip") {
		t.Fatal("should be blocked after 3 failures")
	}

	// Window elapses → key is allowed again.
	now = now.Add(2 * time.Minute)
	if !l.Allow("ip") {
		t.Fatal("should be allowed after window expiry")
	}

	// A success clears the counter immediately.
	l.Fail("ip")
	l.Fail("ip")
	l.Fail("ip")
	if l.Allow("ip") {
		t.Fatal("should be blocked again")
	}
	l.Reset("ip")
	if !l.Allow("ip") {
		t.Fatal("Reset should clear the block")
	}
}
