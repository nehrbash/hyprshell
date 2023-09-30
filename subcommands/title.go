package subcommands

// import (
// 	"context"
// 	"flag"
// 	"fmt"
// 	"os"

// 	core "github.com/nehrbash/hyprshell/pkg"

// 	"github.com/godbus/dbus/v5"
// 	"github.com/google/subcommands"
// )

// // Title is a comannd manager struct
// type Title struct{}

// func (*Title) Name() string     { return "title" }
// func (*Title) Synopsis() string { return "current app title" }
// func (*Title) Usage() string {
// 	return `title
// get current app title as json data: Stream
// `
// }

// // SetFlags adds the check flags to the specified set.
// func (m *Title) SetFlags(f *flag.FlagSet) {
// }

// // Execute executes the check command.
// func (m *Title) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
// 	conn, err := dbus.SessionBus()
// 	if err != nil {
// 		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
// 		return subcommands.ExitSuccess
// 	}

// 	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
// 		fmt.Sprintf("type='signal',sender='%s',interface='%s',path='%s',member='%s'",
// 			core.ServiceName, core.TitleInterfaceName, core.ServiceObjectPath, m.Name()))

// 	signalChan := make(chan *dbus.Signal, 10)
// 	conn.Signal(signalChan)

// 	fmt.Println("{ \"submap\": \"Default\" }")
// 	for signal := range signalChan {
// 		if signal.Name == core.TitleInterfaceName+"."+m.Name() {
// 			message := signal.Body[0].(string)
// 			fmt.Println(message)
// 		}
// 	}
// 	return subcommands.ExitSuccess
// }
