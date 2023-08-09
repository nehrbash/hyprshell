package subcommands

import (
	"context"
	"flag"
	"fmt"
	core "hyprshell/pkg"
	"log"
	"os"

	"github.com/godbus/dbus/v5"
	"github.com/google/subcommands"
)

type Dock struct {
	save bool
}

func (*Dock) Name() string     { return "dock" }
func (*Dock) Synopsis() string { return "print dock information in json in a constant steam" }
func (*Dock) Usage() string {
	return `dock
get dock json data: Stream
`
}

// SetFlags adds the check flags to the specified set.
func (m *Dock) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&m.save, "save", false, "save apps to favorites")
}

// Execute executes the check command.
func (m *Dock) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	conn, err := dbus.SessionBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		return subcommands.ExitSuccess
	}

	if m.save {
		obj := conn.Object(core.ServiceName, core.ServiceObjectPath)
		err = obj.Call(core.ServiceInterface+".DoAction", 0, "save").Store()
		if err != nil {
			log.Print(err)
		}

	} else { // default stream data
		conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
			fmt.Sprintf("type='signal',sender='%s',interface='%s',path='%s',member='%s'",
				core.ServiceName, core.ServiceInterface, core.ServiceObjectPath, m.Name()))

		signalChan := make(chan *dbus.Signal, 1)
		conn.Signal(signalChan)

		for signal := range signalChan {
			if signal.Name == core.ServiceInterface+"."+m.Name() {
				message := signal.Body[0].(string)
				fmt.Println(message)
			}
		}
	}
	return subcommands.ExitSuccess
}
