package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type StatusFn = func(ch chan<- ModuleStatus)

// type StatusFn = func(ch chan<- string, done chan struct{})

type ModuleConfig interface {
	Decode(map[string]interface{}) error
	MakeStatusFn() StatusFn
}

type AppConfig struct {
	Modules []ModuleConfig
	// to be extended, probably
}

type ModuleStatus interface {
	String() string
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

	// TODO: brauche keine done channel, wenn jedes Moudule einen eigenen Channel
	// nutzt. Channel kann einfach geschlossen werden.

	// var wg sync.WaitGroup
	// wg.Add(len(statusFns))
	type Status struct {
		id     int
		status ModuleStatus
	}
	sinkCh := make(chan Status)
	// done := make(chan struct{})
	moduleChans := make([]chan ModuleStatus, len(statusFns))
	for i := range moduleChans {
		moduleChans[i] = make(chan ModuleStatus)
	}
	for i, fn := range statusFns {
		// Annotate any received module status with its respective channel's
		// index.
		go func(j int) {
			for v := range moduleChans[j] {
				sinkCh <- Status{id: j, status: v}
			}
		}(i)
		go fn(moduleChans[i])
		// TODO: wg.Done()
	}

	// ch := make(chan Status)
	// for i := 0; i < len(statusFns); i++ {
	// 	go func(n int) {
	// 		defer wg.Done()
	// 		statusFns[n](n, ch, done)
	// 	}(i)
	// }

	status := make([]string, len(statusFns))
	for {
		st := <-sinkCh
		log.Printf("%d: %T %v\n", st.id, st.status, st.status)

		status[st.id] = st.status.String()
		s := strings.Join(status, " ")
		err := setXRootName(conn, screen, s)
		if err != nil {
			log.Printf("setXRootName: %v", err)
		}
	}
}

func getModuleConfigFromTypeString(t string) (ModuleConfig, error) {
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
	case "cpu":
		return &CpuConfig{}, nil
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

		moduleConfig, err := getModuleConfigFromTypeString(t)
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
