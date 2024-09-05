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

// TODO: Nutzloser Typ? Wird nur zum Unmarshalling verwendet.
type AppCfg = []interface{}

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

	// TODO: ist das nicht völlig unnötig? Kann ich nicht einfach
	// innerhalb von parseConfig() das Array aus Statusfns initialisieren?
	appCfg, err := parseConfig(cfgJson)
	if err != nil {
		log.Fatalf("error parsing config file %s: %v", cfgFilePath, err)
	}

	statusFns := []StatusFn{}
	for _, cfg := range appCfg {
		switch c := cfg.(type) {
		case TimeConfig:
			statusFns = append(statusFns, MakeTimeStatusFn(c))
		case DateConfig:
			statusFns = append(statusFns, MakeDateStatusFn(c))
		case NetspeedConfig:
			statusFns = append(statusFns, MakeNetspeedStatusFn(c))
		case MemInfoConfig:
			statusFns = append(statusFns, MakeMemInfoStatusFn(c))
		case BatteryConfig:
			statusFns = append(statusFns, MakeBatteryStatusFn(c))
		}
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

func parseConfig(cfg []byte) (AppCfg, error) {
	var cfgRaw map[string]interface{}

	if err := json.Unmarshal(cfg, &cfgRaw); err != nil {
		return nil, err
	}

	statusArr, ok := cfgRaw["status"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid status field")
	}

	appCfg := AppCfg{}
	for _, v := range statusArr {
		status, ok := v.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid status field")
		}

		t, ok := status["type"].(string)
		if !ok {
			return nil, fmt.Errorf("type missing in status config")
		}
		switch t {
		case "time":
			timeCfg, err := NewTimeConfig(status)
			if err != nil {
				return nil, fmt.Errorf("error parsing time config: %v", err)
			}
			appCfg = append(appCfg, timeCfg)
		case "date":
			dateCfg, err := NewDateConfig(status)
			if err != nil {
				return nil, fmt.Errorf("error parsing date config: %v", err)
			}
			appCfg = append(appCfg, dateCfg)
		case "netspeed":
			netspeedCfg, err := NewNetspeedConfig(status)
			if err != nil {
				return nil, fmt.Errorf("error parsing netspeed config: %v", err)
			}
			appCfg = append(appCfg, netspeedCfg)
		case "ram":
			ramCfg, err := NewMemInfoConfig(status)
			if err != nil {
				return nil, fmt.Errorf("error parsing RAM config: %v", err)
			}
			appCfg = append(appCfg, ramCfg)
		case "battery":
			batteryCfg, err := NewBatteryConfig(status)
			if err != nil {
				return nil, fmt.Errorf("error parsing battery config: %v", err)
			}
			appCfg = append(appCfg, batteryCfg)
		default:
			return nil, fmt.Errorf("invalid type %s", t)
		}
	}

	return appCfg, nil
}
