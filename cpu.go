package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type CpuConfig struct {
	Period time.Duration
}

func (c *CpuConfig) Decode(m map[string]interface{}) error {
	periodMsF, ok := m["period_ms"].(float64)
	periodMs := int(periodMsF)
	if !ok || periodMs < 1 {
		return fmt.Errorf("invalid period in cpu config")
	}
	c.Period = time.Duration(periodMs) * time.Millisecond

	return nil
}

type CpuStatus struct {
	Usage float64 // CPU usage in percent
}

func (s CpuStatus) String() string {
	return fmt.Sprintf("Cpu %03d%%", int64(s.Usage * 100))
}

// Returns total ticks and idle ticks from /proc/stat
func getCpuTimings() (int64, int64, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	s := bufio.NewScanner(f)
	if !s.Scan() {
		return 0, 0, err
	}
	f.Close()
	line := s.Text()

	splits := strings.Fields(line)
	var total int64
	for _, s := range splits[1:] /* skip "cpu" prefix */ {
		if len(s) == 0 {
			continue
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, 0, err
		}
		total += n
	}
	idle, err := strconv.ParseInt(splits[4], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return total, idle, err
}

func (c CpuConfig) MakeStatusFn() StatusFn {
	return func(ch chan<- ModuleStatus) {
		totalOld, idleOld, err := getCpuTimings()
		if err != nil {
			log.Printf("getCpuTimings error: %v", err)
			totalOld = 0
			idleOld = 0
		}

		tick := time.NewTicker(c.Period)
		for range tick.C {
			totalNew, idleNew, err := getCpuTimings()
			if err != nil {
				log.Printf("getCpuTimings error: %v", err)
				totalOld = 0
				idleOld = 0
			}
			usage := (1.0 - (float64(idleNew-idleOld) / float64(totalNew-totalOld)))
			ch <- CpuStatus{Usage: usage}

			totalOld = totalNew
			idleOld = idleNew
		}
	}
}
