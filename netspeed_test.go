package main

import (
	"testing"
)

func TestNewNetspeedConfig(t *testing.T) {
	table := []struct {
		m     map[string]interface{}
		valid bool
	}{
		{map[string]interface{}{"device": "wlan0", "period_ms": 1000.0}, true},
		{map[string]interface{}{"device": "wlan0", "period_ms": 1.0}, true},
		{map[string]interface{}{"device": "", "period_ms": 1000.0}, false},
		{map[string]interface{}{"device": "wlan0", "period_ms": 0.0}, false},
		{map[string]interface{}{"period_ms": 0.0}, false},
		{map[string]interface{}{"device": "wlan0"}, false},
		{map[string]interface{}{"device": "wlan0", "period_ms": -1.0}, false},
	}

	for ii, input := range table {
		var c NetspeedConfig
		err := c.FromMap(input.m)
		if (err == nil) != input.valid {
			t.Errorf("%d: %s error: %v", ii, input.m, err)
		}
	}
}
