package subcommands

import (
	"context"
	"flag"
	"fmt"
	core "hyprshell/pkg"
	"os"

	"github.com/godbus/dbus/v5"
	"github.com/google/subcommands"
)

// Submap is a comannd manager struct
type Monitor struct{}

func (*Monitor) Name() string     { return "monitor" }
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
			core.ServiceName, core.MonitorInterfaceName, core.ServiceObjectPath, m.Name()))

	signalChan := make(chan *dbus.Signal, 10)
	conn.Signal(signalChan)

	for signal := range signalChan {
		if signal.Name == core.MonitorInterfaceName+"."+m.Name() {
			message := signal.Body[0].(string)
			fmt.Println(message)
		}
	}
	return subcommands.ExitSuccess
}
