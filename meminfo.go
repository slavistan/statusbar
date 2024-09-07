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

type MemStatus struct {
	Total int64 // Total RAM in bytes
	Free  int64 // Available RAM in bytes
}

func (c *MemConfig) Decode(m map[string]interface{}) error {
	periodMsF, ok := m["period_ms"].(float64)
	periodMs := int(periodMsF)
	if !ok || periodMs < 1 {
		return fmt.Errorf("invalid period in time config")
	}
	c.Period = time.Duration(periodMs) * time.Millisecond
	return nil
}

func (m MemStatus) String() string {
	usagePct := int((1.0 - (float64(m.Free) / float64(m.Total))) * 100.0)
	return fmt.Sprintf("Mem % 2d%%", usagePct)
}

// ReadMemInfo parses /proc/meminfo and returns relevant information
// in a MemInfo.
func ReadMemInfo() (MemStatus, error) {
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

func (c MemConfig) MakeStatusFn() StatusFn {
	return func(ch chan<- ModuleStatus) {
		// get := func() ModuleStatus {
		// 	meminfo, err := ReadMemInfo()
		// 	if err != nil {
		// 		log.Printf("ReadMemInfo error: %v", err)
		// 		// TODO: Wie kann ich Statusupdates verhindern, falls ein
		// 		// Fehler auftritt und nur logs ausgeben? Passt hier mit den
		// 		// Abstraktionen nicht zusammen.
		// 		return err.Error()
		// 	}

		// 	return fmt.Sprint(meminfo)
		// }

		tick := time.NewTicker(c.Period)
		defer tick.Stop()

		// ch <- get()
		// LOOP:
		for {
			select {
			case <-tick.C:
				meminfo, err := ReadMemInfo()
				if err != nil {
					log.Println("ReadMemInfo error: %v", err)
				} else {
					ch <- meminfo
				}
				// case <-done:
				// 	break LOOP
			}
		}
	}
}
