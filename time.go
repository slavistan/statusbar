package main

import (
	"fmt"
	"time"
)

type TimeConfig struct {
	Period time.Duration
}

type TimeStatus struct {
	Time time.Time
}

// Returns a TimeConfig from a string map as returned from parsing the json
// config.
func (c *TimeConfig) Decode(m map[string]interface{}) error {
	periodMsF, ok := m["period_ms"].(float64)
	periodMs := int(periodMsF)
	if !ok || periodMs < 1 {
		return fmt.Errorf("invalid period in time config")
	}
	c.Period = time.Duration(periodMs) * time.Millisecond
	return nil
}

func MakeTimeStatusFn(cfg TimeConfig) StatusFn {
	return func(id int, ch chan<- Status, done chan struct{}) {
		fn := func(t time.Time) Status {
			return Status{id: id, status: t.Format("⌚ 15:04:05")}
		}

		tick := time.NewTicker(cfg.Period)
		defer tick.Stop()

		ch <- fn(time.Now())
	LOOP:
		for {
			select {
			case t := <-tick.C:
				ch <- fn(t)
			case <-done:
				break LOOP
			}
		}
	}
}
