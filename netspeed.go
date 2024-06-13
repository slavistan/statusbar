package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"
)

func readRxTxBytes(netDevice string) (int64, int64, error) {
	const rxBytesPath = "/sys/class/net/%s/statistics/rx_bytes"
	const txBytesPath = "/sys/class/net/%s/statistics/tx_bytes"

	content, err := os.ReadFile(fmt.Sprintf(rxBytesPath, netDevice))
	if err != nil {
		return 0, 0, err
	}
	rx, err := strconv.ParseInt(string(content[:len(content)-1]), 10, 64)
	if err != nil {
		return 0, 0, err
	}

	content, err = os.ReadFile(fmt.Sprintf(txBytesPath, netDevice))
	if err != nil {
		return 0, 0, err
	}
	tx, err := strconv.ParseInt(string(content[:len(content)-1]), 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return rx, tx, nil
}

func format(netDevice string, rxBPS float64, txBPS float64) string {
	rxUnit := " "
	switch {
	case rxBPS >= 10000 && rxBPS < 10000000:
		rxUnit = "K"
		rxBPS /= 1000
	case rxBPS >= 10000000:
		rxUnit = "M"
		rxBPS /= 1000000
	}

	txUnit := " "
	switch {
	case txBPS >= 10000 && txBPS < 10000000:
		txUnit = "K"
		txBPS /= 1000
	case txBPS >= 10000000:
		txUnit = "M"
		txBPS /= 1000000
	}
	return fmt.Sprintf("🖧 %s: ⬆️ %04d "+txUnit+"B/s ⬇️ %04d "+rxUnit+"B/s", netDevice, int(txBPS), int(rxBPS))
}

type NetspeedConfig struct {
	Device   string
	PeriodMs int
}

func (c *NetspeedConfig) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	device, ok := raw["device"].(string)
	if !ok || len(device) == 0 {
		return fmt.Errorf("expected 'device' to be non-empty string")
	}

	periodMs, ok := raw["period_ms"].(float64) // JSON numbers are float64 by default
	if !ok || periodMs <= 0 {
		return fmt.Errorf("expected 'period_ms' to be a number greater than 0")
	}

	c.Device = device
	c.PeriodMs = int(math.Ceil(periodMs))
	return nil
}

func MakeStatusNetspeedFn(cfg NetspeedConfig) func(id int, ch chan<- Status, done chan struct{}) {
	return func(id int, ch chan<- Status, done chan struct{}) {
		rxBytesOld, txBytesOld, err := readRxTxBytes(cfg.Device)
		if err != nil {
			log.Println("NetSpeed: ", err.Error())
			ch <- Status{id: id, status: err.Error()}
		}
		timeOld := time.Now()

		ch <- Status{id, format(cfg.Device, 0, 0)}

		tick := time.NewTicker(time.Duration(cfg.PeriodMs))
		defer tick.Stop()

	LOOP:
		for {
			select {
			case <-tick.C:
				rxBytes, txBytes, err := readRxTxBytes(cfg.Device)
				if err != nil {
					log.Println("NetSpeed: ", err.Error())
					ch <- Status{id: id, status: err.Error()}
				}

				// We take a fresh timestamp instead of hardcoding the ticker's
				// duration, as we could be delayed by the write to ch.
				time := time.Now()

				durSec := time.Sub(timeOld).Seconds()
				rxBPS := float64(rxBytes-rxBytesOld) / durSec
				txBPS := float64(txBytes-txBytesOld) / durSec
				ch <- Status{id, format(cfg.Device, rxBPS, txBPS)}

				rxBytesOld = rxBytes
				txBytesOld = txBytes
				timeOld = time

			case <-done:
				break LOOP
			}
		}
	}
}
