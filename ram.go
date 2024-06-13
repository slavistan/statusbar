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

type RAM struct {
	total int64 // Total RAM in bytes
	free  int64 // Available RAM in bytes
}

func ReadRAM() (RAM, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return RAM{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	ram := RAM{}
	re := regexp.MustCompile(`[0-9]+`)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "MemTotal"):
			s := re.FindString(line[len("MemTotal"):])
			memKb, err := strconv.ParseInt(s, 10, 64)
			if err != nil || memKb == 0 {
				return RAM{}, errors.New("parsing /proc/meminfo failed")
			}
			ram.total = memKb * 1000
		case strings.HasPrefix(scanner.Text(), "MemAvailable"):
			s := re.FindString(line[len("MemAvailable"):])
			memKb, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return RAM{}, errors.New("parsing /proc/meminfo failed")
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

func StatusRAMUsage(id int, ch chan<- Status, done chan struct{}) {
	fn := func() Status {
		ram, err := ReadRAM()
		if err != nil {
			log.Println("RAM usage: ", err.Error())
			return Status{id: id, status: err.Error()}
		}

		usagePct := int((1.0 - (float64(ram.free) / float64(ram.total))) * 100.0)
		return Status{id: id, status: fmt.Sprintf("RAM %02d%%", usagePct)}
	}

	tick := time.NewTicker(time.Second)
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
