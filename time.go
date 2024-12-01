package main

import (
	"fmt"
	"time"
)

type TimeConfig struct {
	Period time.Duration
}

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
		// TODO: reagiert nicht auf Änderung der Zeitzone
		for t := range tick.C {
			ch <- TimeStatus(t)
		}
	}
}

type TimeStatus time.Time

func (s TimeStatus) String() string {
	return time.Time(s).Format("⌚ 15:04:05")
}
