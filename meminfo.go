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

type MemInfoConfig struct {
	Period time.Duration
}

type MemInfo struct {
	total int64 // Total RAM in bytes
	free  int64 // Available RAM in bytes
}

func NewMemInfoConfig(m map[string]interface{}) (MemInfoConfig, error) {
	periodMsF, ok := m["period_ms"].(float64)
	periodMs := int(periodMsF)
	if !ok || periodMs < 1 {
		return MemInfoConfig{}, fmt.Errorf("invalid period in time config")
	}
	return MemInfoConfig{Period: time.Duration(periodMs) * time.Millisecond}, nil
}

// TODO: Sollten alle Statusmodule einfach String()able sein
// und innerhalb der main() loop ihre Stringrepräsentation über fmt.Print() erhalten?
func (m MemInfo) String() string {
	usagePct := int((1.0 - (float64(m.free) / float64(m.total))) * 100.0)
	return fmt.Sprintf("Mem % 2d%%", usagePct)
}

// ReadMemInfo parses /proc/meminfo and returns relevant information
// in a MemInfo.
func ReadMemInfo() (MemInfo, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return MemInfo{}, fmt.Errorf("error reading /proc/meminfo: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	ram := MemInfo{}
	re := regexp.MustCompile(`[0-9]+`)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "MemTotal"):
			s := re.FindString(line[len("MemTotal"):])
			memKb, err := strconv.ParseInt(s, 10, 64)
			if err != nil || memKb == 0 {
				return MemInfo{}, errors.New("parsing /proc/meminfo failed")
			}
			ram.total = memKb * 1000
		case strings.HasPrefix(scanner.Text(), "MemAvailable"):
			s := re.FindString(line[len("MemAvailable"):])
			memKb, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return MemInfo{}, errors.New("parsing /proc/meminfo failed")
			}
			ram.free = memKb * 1000
		}

		// Exit early after values of interest have been read.
		if ram.free != 0 && ram.total != 0 {
			break
		}
	}

	return ram, nil
}

func MakeMemInfoStatusFn(cfg MemInfoConfig) StatusFn {
	return func(id int, ch chan<- Status, done chan struct{}) {
		fn := func() Status {
			meminfo, err := ReadMemInfo()
			if err != nil {
				log.Printf("ReadMemInfo error: %v", err)
				// TODO: Wie kann ich Statusupdates verhindern, falls ein
				// Fehler auftritt und nur logs ausgeben? Passt hier mit den
				// Abstraktionen nicht zusammen.
				return Status{id: id, status: err.Error()}
			}

			// usagePct := int((1.0 - (float64(ram.free) / float64(ram.total))) * 100.0)
			return Status{id: id, status: fmt.Sprint(meminfo)}
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
