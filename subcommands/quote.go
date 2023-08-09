package subcommands

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	core "github.com/nehrbash/hyprshell/pkg"

	"github.com/godbus/dbus/v5"
	"github.com/google/subcommands"
)

// Submap is a comannd manager struct
type Quote struct{}

func (*Quote) Name() string     { return "quote" }
func (*Quote) Synopsis() string { return "json quote with author" }
func (*Quote) Usage() string {
	return `quote
json Quote and Author
`
}

// SetFlags adds the check flags to the specified set.
func (m *Quote) SetFlags(f *flag.FlagSet) {
}

// Execute executes the check command.
func (m *Quote) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	conn, err := dbus.SessionBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		return subcommands.ExitSuccess
	}

	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		fmt.Sprintf("type='signal',sender='%s',interface='%s',path='%s',member='%s'",
			core.ServiceName, core.QuoteInterfaceName, core.ServiceObjectPath, "quote"))

	signalChan := make(chan *dbus.Signal, 10)
	conn.Signal(signalChan)

	// request first quote
	obj := conn.Object(core.ServiceName, core.ServiceObjectPath)
	err = obj.Call(core.QuoteInterfaceName+".GetQuote", 0, "get").Store()
	if err != nil {
		log.Print(err)
	}

	for signal := range signalChan {
		if signal.Name == core.QuoteInterfaceName+"."+m.Name() {
			message := signal.Body[0].(string)
			fmt.Println(message)
		}
	}
	return subcommands.ExitSuccess
}
