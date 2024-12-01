package main

import (
	"fmt"
	"time"
)

type DateConfig struct {
	Period time.Duration
}

type DateStatus time.Time

func (d DateStatus) String() string {
	return time.Time(d).Format("📅 2006-01-02")
}

func (c *DateConfig) FromMap(m map[string]interface{}) error {
	periodMsF, ok := m["period_ms"].(float64)
	periodMs := int(periodMsF)
	if !ok || periodMs < 1 {
		return fmt.Errorf("invalid period in date config")
	}
	c.Period = time.Duration(periodMs) * time.Millisecond
	return nil
}

func (c DateConfig) MakeStatusFn() StatusFn {
	return func(ch chan<- ModuleStatus) {
		tick := time.NewTicker(c.Period)
		defer tick.Stop()

		ch <- DateStatus(time.Now())
		// TODO: reagiert nicht auf Änderung der Zeitzone
		for t := range tick.C {
			ch <- DateStatus(t)
		}
	}
}
