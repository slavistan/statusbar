package main

import (
	"fmt"
	"time"
)

type DateConfig struct {
	Period time.Duration
}

// Returns a DateConfig from a string map as returned from parsing the json
// config.
func NewDateConfig(m map[string]interface{}) (DateConfig, error) {
	periodMsF, ok := m["period_ms"].(float64)
	periodMs := int(periodMsF)
	if !ok || periodMs < 1 {
		return DateConfig{}, fmt.Errorf("invalid period in date config")
	}

	return DateConfig{Period: time.Millisecond * time.Duration(periodMs)}, nil
}

func MakeDateStatusFn(cfg DateConfig) StatusFn {
	return func(id int, ch chan<- Status, done chan struct{}) {
		fn := func(t time.Time) Status {
			return Status{id: id, status: t.Format("📅 2006-01-02")}
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
