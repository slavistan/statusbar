// See https://www.kernel.org/doc/html/latest/power/power_supply_class.html.

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type BatteryConfig struct {
	Battery string
	Period  time.Duration
}

func (c *BatteryConfig) FromMap(m map[string]interface{}) error {
	periodMsF, ok := m["period_ms"].(float64)
	periodMs := int(periodMsF)
	if !ok || periodMs < 1 {
		return fmt.Errorf("invalid period in battery config")
	}
	c.Period = time.Duration(periodMs) * time.Millisecond

	battery, ok := m["battery"].(string)
	if !ok {
		return fmt.Errorf("invalid battery in battery config")
	}
	c.Battery = battery

	return nil
}

func (c BatteryConfig) MakeStatusFn() StatusFn {
	return func(ch chan<- ModuleStatus) {
		tick := time.NewTicker(c.Period)
		defer tick.Stop()
		for range tick.C {
			bat, err := readBattery(c.Battery)
			if err != nil {
				log.Printf("ReadBattery error: %v", err)
			} else {
				ch <- bat
			}
		}
	}
}

type BatteryStatus struct {
	Capacity int  // battery capacity in percent
	ACOnline bool // whether the AC is connected
}

func (b BatteryStatus) String() string {
	var c string
	if b.ACOnline {
		c = "🔌"
	} else {
		c = "🔋"
	}
	return fmt.Sprintf("%s %03d%%", c, b.Capacity)
}

func readBattery(battery string) (BatteryStatus, error) {
	const POWER_SUPPLY_CAPACITY = "/sys/class/power_supply/%s/capacity"
	const POWER_SUPPLY_AC_ONLINE = "/sys/class/power_supply/AC/online"

	p := fmt.Sprintf(POWER_SUPPLY_CAPACITY, battery)
	capStr, err := os.ReadFile(p)
	if err != nil {
		return BatteryStatus{}, fmt.Errorf("failed to read %s: %v", p, err)
	}
	capStr = capStr[:len(capStr)-1] // remove trailing newline

	acOnlineStr, err := os.ReadFile(POWER_SUPPLY_AC_ONLINE)
	if err != nil {
		return BatteryStatus{}, fmt.Errorf("failed to read %s: %v", p, err)
	}
	acOnlineStr = acOnlineStr[:len(acOnlineStr)-1] // remove trailing newline

	cap, err := strconv.Atoi(string(capStr))
	if err != nil {
		return BatteryStatus{}, fmt.Errorf("failed to atoi %s: %v", capStr, err)
	}
	acOnline := string(acOnlineStr) != "0"
	return BatteryStatus{cap, acOnline}, nil
}
