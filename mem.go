package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type MemConfig struct {
	Period time.Duration
}

func (c *MemConfig) FromMap(m map[string]interface{}) error {
	periodMsF, ok := m["period_ms"].(float64)
	periodMs := int(periodMsF)
	if !ok || periodMs < 1 {
		return fmt.Errorf("invalid period in time config")
	}
	c.Period = time.Duration(periodMs) * time.Millisecond
	return nil
}

func (c MemConfig) MakeStatusFn() StatusFn {
	return func(ch chan<- ModuleStatus) {

		tick := time.NewTicker(c.Period)
		defer tick.Stop()

		for range tick.C {
			meminfo, err := readMemInfo()
			if err != nil {
				log.Printf("mem: %v", err)
			} else {
				ch <- meminfo
			}
		}
	}
}

type MemStatus struct {
	Total int64 // Total RAM in bytes
	Free  int64 // Available RAM in bytes
}

func (m MemStatus) String() string {
	usagePct := int((1.0 - (float64(m.Free) / float64(m.Total))) * 100.0)
	return fmt.Sprintf("Mem %03d%%", usagePct)
}

// readMemInfo parses /proc/meminfo and returns relevant information
// in a MemStatus.
func readMemInfo() (MemStatus, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return MemStatus{}, fmt.Errorf("error reading /proc/meminfo: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	ram := MemStatus{}
	re := regexp.MustCompile(`[0-9]+`)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "MemTotal"):
			s := re.FindString(line[len("MemTotal"):])
			memKb, err := strconv.ParseInt(s, 10, 64)
			if err != nil || memKb == 0 {
				return MemStatus{}, errors.New("parsing /proc/meminfo failed")
			}
			ram.Total = memKb * 1000
		case strings.HasPrefix(scanner.Text(), "MemAvailable"):
			s := re.FindString(line[len("MemAvailable"):])
			memKb, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return MemStatus{}, errors.New("parsing /proc/meminfo failed")
			}
			ram.Free = memKb * 1000
		}

		// Exit early after values of interest have been read.
		if ram.Free != 0 && ram.Total != 0 {
			break
		}
	}

	return ram, nil
}
