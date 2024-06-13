package main

import (
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"log"
	"strings"
	"sync"
	"time"
)

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

	// List status functions. Their statuses will be displayed in the order
	// defined here.
	statusFns := []func(id int, ch chan<- Status, done chan struct{}){
		MakeStatusNetspeedFn("enp69s0"),
		StatusRAMUsage,
		statusDate,
		statusTime,
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

func statusDate(id int, ch chan<- Status, done chan struct{}) {
	fn := func(t time.Time) Status {
		return Status{id: id, status: t.Format("📅 2006-01-02")}
	}

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	ch <- fn(time.Now())
LOOP:
	for {
		select {
		case t := <-tick.C:
			ch <- fn(t)
		case <-done:
			break LOOP
		}
	}
}

func statusTime(id int, ch chan<- Status, done chan struct{}) {
	fn := func(t time.Time) Status {
		return Status{id: id, status: t.Format("⌚ 15:04:05")}
	}

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	ch <- fn(time.Now())
LOOP:
	for {
		select {
		case t := <-tick.C:
			ch <- fn(t)
		case <-done:
			break LOOP
		}
	}
}
