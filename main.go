package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type StatusFn = func(id int, ch chan<- Status, done chan struct{})

type ModuleConfig interface {
	Decode(map[string]interface{}) error
	MakeStatusFn() StatusFn
}

type AppConfig struct {
	Modules []ModuleConfig
	// to be extended, probably
}

type Status struct {
	id     int
	status string
}

func setXRootName(conn *xgb.Conn, screen *xproto.ScreenInfo, v string) error {
	return xproto.ChangePropertyChecked(
		conn,
		xproto.PropModeReplace,
		screen.Root,
		xproto.AtomWmName,
		xproto.AtomString,
		8,
		uint32(len(v)),
		[]byte(v),
	).Check()
}

func main() {
	// Init X11 connection
	conn, err := xgb.NewConn()
	if err != nil {
		log.Fatal("cannot initiate X11 connection")
	}
	screen := xproto.Setup(conn).DefaultScreen(conn)
	defer conn.Close()

	cfgFilePath := "config.json"
	cfgJson, err := os.ReadFile(cfgFilePath)
	if err != nil {
		log.Fatalf("error reading config file %s: %v", cfgFilePath, err)
	}

	var appCfg AppConfig
	json.Unmarshal(cfgJson, &appCfg)
	if err != nil {
		log.Fatalf("error parsing config file %s: %v", cfgFilePath, err)
	}
	statusFns := []StatusFn{}
	for _, cfg := range appCfg.Modules {
		statusFns = append(statusFns, cfg.MakeStatusFn())
	}

	var wg sync.WaitGroup
	wg.Add(len(statusFns))
	ch := make(chan Status)
	done := make(chan struct{})
	for i := 0; i < len(statusFns); i++ {
		go func(n int) {
			defer wg.Done()
			statusFns[n](n, ch, done)
		}(i)
	}

	status := make([]string, len(statusFns))
	for {
		st := <-ch
		log.Printf("%v\n", st)

		status[st.id] = st.status
		s := strings.Join(status, " ")
		err := setXRootName(conn, screen, s)
		if err != nil {
			log.Printf("updating X11 root name failed: %v", err)
		}
	}
}

func getConfigFromTypeString(t string) (ModuleConfig, error) {
	switch t {
	case "time":
		return &TimeConfig{}, nil
	case "date":
		return &DateConfig{}, nil
	case "netspeed":
		return &NetspeedConfig{}, nil
	case "ram":
		return &MemConfig{}, nil
	case "battery":
		return &BatteryConfig{}, nil
	default:
		return nil, fmt.Errorf("invalid type %s", t)
	}
}

func (c *AppConfig) UnmarshalJSON(data []byte) error {
	var cfgRaw map[string]interface{}
	if err := json.Unmarshal(data, &cfgRaw); err != nil {
		return nil
	}

	statusArr, ok := cfgRaw["status"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid status field")
	}

	c.Modules = []ModuleConfig{}
	for _, v := range statusArr {
		status, ok := v.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid status field")
		}

		t, ok := status["type"].(string)
		if !ok {
			return fmt.Errorf("type missing in status config")
		}

		moduleConfig, err := getConfigFromTypeString(t)
		if err != nil {
			return err
		}
		if err := moduleConfig.Decode(status); err != nil {
			return fmt.Errorf("error decoding %s config: %v", t, err)
		}
		c.Modules = append(c.Modules, moduleConfig)
	}
	return nil
}
