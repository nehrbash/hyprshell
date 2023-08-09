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
type Submap struct{}

func (*Submap) Name() string     { return "submap" }
func (*Submap) Synopsis() string { return "print hyprland supmap in json as a constant steam" }
func (*Submap) Usage() string {
	return `submap
get hyprland mode as json data: Stream
`
}

// SetFlags adds the check flags to the specified set.
func (m *Submap) SetFlags(f *flag.FlagSet) {
}

// Execute executes the check command.
func (m *Submap) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	conn, err := dbus.SessionBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		return subcommands.ExitSuccess
	}

	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		fmt.Sprintf("type='signal',sender='%s',interface='%s',path='%s',member='%s'",
			core.ServiceName, core.SubmapInterfaceName, core.ServiceObjectPath, m.Name()))

	signalChan := make(chan *dbus.Signal, 10)
	conn.Signal(signalChan)

	fmt.Println("{ \"submap\": \"Default\" }")
	for signal := range signalChan {
		if signal.Name == core.SubmapInterfaceName+"."+m.Name() {
			message := signal.Body[0].(string)
			fmt.Println(message)
		}
	}
	return subcommands.ExitSuccess
}
