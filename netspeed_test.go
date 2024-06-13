package main

import (
	"testing"
)

func TestNetspeedUnmarshalling(t *testing.T) {
	table := []struct {
		json  string
		valid bool
	}{
		{`{ "device": "wlan0", "period_ms": 1000 }`, true},
		{`{ "device": "wlan0", "period_ms": 1 }`, true},
		{`{ "device": "", "period_ms": 1000 }`, false},
		{`{ "device": "wlan0", "period_ms": 0 }`, false},
		{`{ "period_ms": 0 }`, false},
		{`{ "device": "wlan0" }`, false},
		{`{ "device": "wlan0", "period_ms": -1 }`, false},
	}

	for _, input := range table {
		var c NetspeedConfig
		err := c.UnmarshalJSON([]byte(input.json))
		if (err == nil) != input.valid {
			t.Errorf("%s error: %v", input.json, err)
		}
	}
}
