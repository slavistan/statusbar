package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type NetspeedConfig struct {
	Device string        // network device name
	Period time.Duration // update period
}

func (c *NetspeedConfig) Decode(m map[string]interface{}) error {
	periodMsF, ok := m["period_ms"].(float64)
	periodMs := int(periodMsF)
	if !ok || periodMs < 1 {
		return fmt.Errorf("invalid period in netspeed config")
	}
	c.Period = time.Duration(periodMs) * time.Millisecond

	device, ok := m["device"].(string)
	if !ok || len(device) == 0 {
		return fmt.Errorf("invalid device in netspeed config")
	}
	c.Device = device

	return nil
}

type NetspeedStatus struct {
	Device  string
	UpBPS   float64
	DownBPS float64
}

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

func (s NetspeedStatus) String() string {
	rxUnit := " "
	switch {
	case s.DownBPS >= 10000 && s.DownBPS < 10000000:
		rxUnit = "K"
		s.DownBPS /= 1000
	case s.DownBPS >= 10000000:
		rxUnit = "M"
		s.DownBPS /= 1000000
	}

	txUnit := " "
	switch {
	case s.UpBPS >= 10000 && s.UpBPS < 10000000:
		txUnit = "K"
		s.UpBPS /= 1000
	case s.UpBPS >= 10000000:
		txUnit = "M"
		s.UpBPS /= 1000000
	}
	return fmt.Sprintf("🖧 %s: ⬆️ %04d "+txUnit+"B/s ⬇️ %04d "+rxUnit+"B/s", s.Device, int(s.UpBPS), int(s.DownBPS))
}

func (c NetspeedConfig) MakeStatusFn() StatusFn {
	return func(ch chan<- ModuleStatus) {
		rxBytesOld, txBytesOld, err := readRxTxBytes(c.Device)
		if err != nil {
			log.Printf("readRxTxByes error: %v", err)
			rxBytesOld = 0
			txBytesOld = 0
		}
		timeOld := time.Now()

		// ch <- NetspeedStatus{c.Device, 0, 0}

		tick := time.NewTicker(time.Duration(c.Period))
		defer tick.Stop()

		// LOOP:
		for {
			select {
			case timeNew := <-tick.C:
				rxBytes, txBytes, err := readRxTxBytes(c.Device)
				if err != nil {
					log.Printf("readRxTxBytes error: %v\n", err)
					rxBytesOld = 0
					txBytesOld = 0
				}

				// We take a fresh timestamp instead of hardcoding the ticker's
				// duration, as we could be delayed by the write to ch.

				durSec := timeNew.Sub(timeOld).Seconds()
				rxBPS := float64(rxBytes-rxBytesOld) / durSec
				txBPS := float64(txBytes-txBytesOld) / durSec
				ch <- NetspeedStatus{c.Device, txBPS, rxBPS}

				rxBytesOld = rxBytes
				txBytesOld = txBytes
				timeOld = timeNew

				// case <-done:
				// 	break LOOP
			}
		}
	}
}
