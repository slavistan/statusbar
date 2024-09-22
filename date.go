package main

// FIXME: Bezieht Änderungen der Zeitzone nicht mit ein (timedatectl set-timezone ...)
// Selbes für time.go

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

func (c *DateConfig) Decode(m map[string]interface{}) error {
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
		get := func(t time.Time) ModuleStatus {
			d := DateStatus(t)
			return d
		}

		tick := time.NewTicker(c.Period)
		defer tick.Stop()

		ch <- get(time.Now())
		// LOOP:
		for {
			select {
			case t := <-tick.C:
				ch <- get(t)
				// case <-done:
				// 	break LOOP
			}
		}
	}
}
