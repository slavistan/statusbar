package main

import (
	"fmt"
	"time"
)

type TimeConfig struct {
	Period time.Duration
}

type TimeStatus time.Time

func (s TimeStatus) String() string {
	return time.Time(s).Format("⌚ 15:04:05")
}

// Returns a TimeConfig from a string map as returned from parsing the json
// config.
func (c *TimeConfig) FromMap(m map[string]interface{}) error {
	periodMsF, ok := m["period_ms"].(float64)
	periodMs := int(periodMsF)
	if !ok || periodMs < 1 {
		return fmt.Errorf("invalid period in time config")
	}
	c.Period = time.Duration(periodMs) * time.Millisecond
	return nil
}

func (c TimeConfig) MakeStatusFn() StatusFn {
	return func(ch chan<- ModuleStatus) {
		tick := time.NewTicker(c.Period)
		defer tick.Stop()

		ch <- TimeStatus(time.Now())
		// LOOP:
		for {
			select {
			case t := <-tick.C:
				ch <- TimeStatus(t)
				// case <-done:
				// 	break LOOP
			}
		}
	}
}
