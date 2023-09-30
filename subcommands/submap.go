package subcommands

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	core "github.com/nehrbash/hyprshell/pkg"

	"github.com/godbus/dbus/v5"
	"github.com/google/subcommands"
)

// Submap is a comannd manager struct
type Submap struct{}

func (*Submap) Name() string { return "submap" }

func (s *Submap) ServiceName() string {
	return core.InterfacePrefix + s.Name()
}

func (s *Submap) InterfaceName() string {
	return core.InterfacePrefix + s.Name() + ".Reciver"
}
func (s *Submap) ServicePath() dbus.ObjectPath {
	return dbus.ObjectPath(path.Join(core.BaseServiceObjectPath, s.Name()))
}
func (s *Submap) ServiceRun(conn *dbus.Conn, msg any) {
	log.Print("submap signal received")

	currentMap := msg.(string)
	if currentMap == "" {
		currentMap = "Default"
	}
	data := fmt.Sprintf(`{ "submap": "%s" }`, currentMap)
	err := conn.Emit(s.ServicePath(), s.InterfaceName(), data)
	if err != nil {
		log.Print(err)
	}

}

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
			m.ServiceName(), m.InterfaceName(), m.ServicePath(), m.Name()))

	signalChan := make(chan *dbus.Signal, 10)
	conn.Signal(signalChan)

	fmt.Println("{ \"submap\": \"Default\" }")
	for signal := range signalChan {
		if signal.Name == m.InterfaceName() {
			message := signal.Body[0].(string)
			fmt.Println(message)
		}
	}
	return subcommands.ExitSuccess
}
