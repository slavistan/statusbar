package main

import (
	"testing"
)

func TestParseConfig(t *testing.T) {
	table := []struct {
		json  string
		valid bool
	}{
		{`{ "status": [{"type": "time", "period_ms": 1000}] }`, true},
		{`{ "err": [{"type": "time", "period_ms": 1000}] }`, false}, // status array missing
		{`{ "status": 2 }`, false},                                  // status is not an array
		{`[1,2,3]`, false},                                          // config is not a json object
		{`{ "err": [{"type": "foo", "period_ms": 1000}] }`, false},  // invalid type foo
	}

	for ii, input := range table {
		_, err := parseConfig([]byte(input.json))
		if (err == nil) != input.valid {
			t.Errorf("%d: %s error: %v", ii, input.json, err)
		}
	}
}
