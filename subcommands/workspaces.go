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
type Workspaces struct{}

func (*Workspaces) Name() string     { return "workspaces" }
func (*Workspaces) Synopsis() string { return "print hyprland monitor index as json" }
func (*Workspaces) Usage() string {
	return `submap
get hyprland monitor indexe as json data: Stream
`
}

// SetFlags adds the check flags to the specified set.
func (m *Workspaces) SetFlags(f *flag.FlagSet) {
}

// Execute executes the check command.
func (m *Workspaces) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	conn, err := dbus.SessionBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		return subcommands.ExitSuccess
	}

	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		fmt.Sprintf("type='signal',sender='%s',interface='%s',path='%s',member='%s'",
			core.ServiceName, core.WorkspacesInterfaceName, core.ServiceObjectPath, m.Name()))

	signalChan := make(chan *dbus.Signal, 10)
	conn.Signal(signalChan)

	for signal := range signalChan {
		if signal.Name == core.WorkspacesInterfaceName+"."+m.Name() {
			message := signal.Body[0].(string)
			fmt.Println(message)
		}
	}
	return subcommands.ExitSuccess
}
