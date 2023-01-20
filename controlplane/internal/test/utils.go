package test

import (
	"time"
)

func AssertWithTimeout(cond func() bool, timeout time.Duration, assertion func()) {
	check := time.NewTicker(time.Duration(200) * time.Millisecond)
	t := time.NewTicker(timeout)
loop:
	for {
		select {
		case <-check.C:
			if cond() {
				break loop
			}
		case <-t.C:
			break loop
		}
	}

	assertion()
}
