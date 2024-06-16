package main

import (
	"fmt"
	"time"
)

type TimeConfig struct {
	Period time.Duration
}

// Returns a TimeConfig from a string map as returned from parsing the json
// config.
func NewTimeConfig(m map[string]interface{}) (TimeConfig, error) {
	periodMsF, ok := m["period_ms"].(float64)
	periodMs := int(periodMsF)
	if !ok || periodMs < 1 {
		return TimeConfig{}, fmt.Errorf("invalid period in time config")
	}

	return TimeConfig{Period: time.Duration(periodMs) * time.Millisecond}, nil
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
