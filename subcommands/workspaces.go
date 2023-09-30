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
type Workspaces struct {
	workspaces   core.Workspaces
	updateSignal chan string
}

func (w *Workspaces) ServiceName() string {
	return core.InterfacePrefix + w.Name()
}

func (w *Workspaces) InterfaceName() string {
	return core.InterfacePrefix + w.Name() + ".Receiver"
}
func (w *Workspaces) ServicePath() dbus.ObjectPath {
	return dbus.ObjectPath(path.Join(core.BaseServiceObjectPath, w.Name()))
}
func (w *Workspaces) ServiceRun(conn *dbus.Conn, msg any) {
	mon := core.GetMonitors()
	w.workspaces.Update(mon)
	data := w.workspaces.String()
	err := conn.Emit(w.ServicePath(), w.InterfaceName()+"."+w.Name(), data)
	if err != nil {
		log.Print(err)
	}

}

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
func (w *Workspaces) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	conn, err := dbus.SessionBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		return subcommands.ExitSuccess
	}

	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		fmt.Sprintf("type='signal',sender='%s',interface='%s',path='%s',member='%s'",
			w.ServiceName(), w.InterfaceName(), w.ServicePath(), w.Name()))

	signalChan := make(chan *dbus.Signal)
	conn.Signal(signalChan)

	// request first quote
	obj := conn.Object(w.ServiceName(), w.ServicePath())
	err = obj.Call("Update", 0).Store()
	if err != nil {
		log.Print(err)
	}

	for signal := range signalChan {
		if signal.Name == w.InterfaceName()+"."+w.Name() {
			message := signal.Body[0].(string)
			fmt.Println(message)
		}
	}
	return subcommands.ExitSuccess
}

func (w *Workspaces) Update() *dbus.Error {
	// TODO why do I only need to do this twice every other time?
	w.updateSignal <- "update plz"
	w.updateSignal <- "update plz"
	return nil
}
