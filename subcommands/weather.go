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

type Weather struct{}

func (*Weather) Name() string     { return "weather" }
func (*Weather) Synopsis() string { return "print weather information in json in a constant steam" }
func (*Weather) Usage() string {
	return `weather
get weather json data: Stream
`
}

// SetFlags adds the check flags to the specified set.
func (w *Weather) SetFlags(f *flag.FlagSet) {}

// Execute executes the check command.
func (w *Weather) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	conn, err := dbus.SessionBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		return subcommands.ExitSuccess
	}

	// setup
	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		fmt.Sprintf("type='signal',sender='%s',interface='%s',path='%s',member='%s'",
			core.ServiceName, core.ServiceInterfaceWeather, core.ServiceObjectPath, "Weather"))
	signalChan := make(chan *dbus.Signal, 10)
	conn.Signal(signalChan)

	// request weather to be sent then listen to the periodic sending
	obj := conn.Object(core.ServiceName, core.ServiceObjectPath)
	err = obj.Call(core.ServiceInterfaceWeather+".GetWeather", 0, "get the weather").Store()
	if err != nil {
		log.Print(err)
	}
	// handle
	for signal := range signalChan {
		if signal.Name == core.ServiceInterfaceWeather+".Weather" {
			fmt.Println(signal.Body[0].(string))
		}
	}

	return subcommands.ExitSuccess
}
