// See https://www.kernel.org/doc/html/latest/power/power_supply_class.html.

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

const POWER_SUPPLY_CAPACITY = "/sys/class/power_supply/%s/capacity"
const POWER_SUPPLY_AC_ONLINE = "/sys/class/power_supply/AC/online"

type BatteryConfig struct {
	Battery string
	Period  time.Duration
}

type Battery struct {
	Capacity int  // battery capacity in percent
	ACOnline bool // whether the AC is connected
}

func (b Battery) String() string {
	var c string
	if b.ACOnline {
		c = "🔌"
	} else {
		c = "🔋"
	}
	return fmt.Sprintf("%s % 3d%%", c, b.Capacity)
}

func ReadBattery(battery string) (Battery, error) {
	p := fmt.Sprintf(POWER_SUPPLY_CAPACITY, battery)
	capStr, err := os.ReadFile(p)
	if err != nil {
		return Battery{}, fmt.Errorf("failed to read %s: %v", p, err)
	}
	capStr = capStr[:len(capStr)-1] // remove trailing newline

	acOnlineStr, err := os.ReadFile(POWER_SUPPLY_AC_ONLINE)
	if err != nil {
		return Battery{}, fmt.Errorf("failed to read %s: %v", p, err)
	}
	acOnlineStr = acOnlineStr[:len(acOnlineStr)-1] // remove trailing newline

	cap, err := strconv.Atoi(string(capStr))
	if err != nil {
		return Battery{}, fmt.Errorf("failed to atoi %s: %v", capStr, err)
	}
	acOnline := string(acOnlineStr) != "0"
	return Battery{cap, acOnline}, nil
}

func NewBatteryConfig(m map[string]interface{}) (BatteryConfig, error) {
	periodMsF, ok := m["period_ms"].(float64)
	periodMs := int(periodMsF)
	if !ok || periodMs < 1 {
		return BatteryConfig{}, fmt.Errorf("invalid period in battery config")
	}
	battery, ok := m["battery"].(string)
	if !ok {
		return BatteryConfig{}, fmt.Errorf("invalid battery in battery config")
	}

	return BatteryConfig{Period: time.Duration(periodMs) * time.Millisecond, Battery: battery}, nil
}

func MakeBatteryStatusFn(cfg BatteryConfig) StatusFn {
	return func(id int, ch chan<- Status, done chan struct{}) {
		fn := func() Status {
			battery, err := ReadBattery(cfg.Battery)
			if err != nil {
				log.Printf("ReadBattery error: %v", err)
				return Status{id, "error"}
			}
			return Status{id: id, status: fmt.Sprint(battery)}
		}

		tick := time.NewTicker(cfg.Period)
		defer tick.Stop()

		ch <- fn()
	LOOP:
		for {
			select {
			case <-tick.C:
				ch <- fn()
			case <-done:
				break LOOP
			}
		}
	}
}
