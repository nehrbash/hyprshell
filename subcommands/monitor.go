package subcommands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/godbus/dbus/v5"
	"github.com/google/subcommands"
	"github.com/joshsziegler/zgo/pkg/log"
	core "github.com/nehrbash/hyprshell/pkg"
)

// Submap is a comannd manager struct
type Monitor struct {
	updateSignal chan string
	monitors     core.Monitors
}

func (*Monitor) Name() string { return "monitor" }
func (m *Monitor) ServiceName() string {
	return core.InterfacePrefix + m.Name()
}

func (m *Monitor) InterfaceName() string {
	return core.InterfacePrefix + m.Name() + ".Reciver"
}

func (m *Monitor) ServicePath() dbus.ObjectPath {
	return dbus.ObjectPath(path.Join(core.BaseServiceObjectPath, m.Name()))
}

type monitor struct {
	ID    int    `json:"id"`
	Model string `json:"model"`
}

func (m *Monitor) ServiceRun(conn *dbus.Conn, msg any) {
	log.Info("focusedmon signal received")
	m.monitors = core.GetMonitors()
	mod := m.monitors.Active()
	moditorData := monitor{
		ID:    mod.ID,
		Model: mod.Model,
	}

	data, err := json.Marshal(moditorData)
	if err != nil {
		log.Info(err)
		return
	}
	err = conn.Emit(m.ServicePath(), m.InterfaceName()+"."+m.Name(), string(data))
	if err != nil {
		log.Info(err)
	}
}

func (*Monitor) Synopsis() string { return "print hyprland monitor index as json" }
func (*Monitor) Usage() string {
	return `submap
get hyprland monitor indexe as json data: Stream
`
}

// SetFlags adds the check flags to the specified set.
func (m *Monitor) SetFlags(f *flag.FlagSet) {
}

// Execute executes the check command.
func (m *Monitor) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	conn, err := dbus.SessionBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		return subcommands.ExitSuccess
	}

	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		fmt.Sprintf("type='signal',sender='%s',interface='%s',path='%s',member='%s'",
			m.ServiceName(), m.InterfaceName(), m.ServicePath(), m.Name()))
	signalChan := make(chan *dbus.Signal)
	conn.Signal(signalChan)

	// request first quote
	obj := conn.Object(m.ServiceName(), m.ServicePath())
	err = obj.Call("Update", 0).Store()
	if err != nil {
		log.Info(err)
	}

	for signal := range signalChan {
		if signal.Name == m.InterfaceName()+"."+m.Name() {
			sig, ok := signal.Body[0].(string)
			if ok {
				fmt.Println(sig)
			}
		}
	}
	return subcommands.ExitSuccess
}

func (m *Monitor) Update() *dbus.Error {
	m.updateSignal <- "request update bitch"
	return nil
}
