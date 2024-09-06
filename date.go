package main

import (
	"fmt"
	"time"
)

type DateConfig struct {
	Period time.Duration
}

type DateStatus struct {
	Date time.Time
}

func (d DateStatus) String() string {
	return d.Date.Format("📅 2006-01-02")
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
	return func(id int, ch chan<- Status, done chan struct{}) {
		fn := func(t time.Time) Status {
			d := DateStatus{Date: t}
			return Status{id: id, status: fmt.Sprint(d)}
		}

		tick := time.NewTicker(c.Period)
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
