package subcommands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	core "github.com/nehrbash/hyprshell/pkg"

	"github.com/godbus/dbus/v5"
	"github.com/google/subcommands"
	"github.com/joshsziegler/zgo/pkg/log"
)

type Title struct {
	updateSignal chan string
	CurrentName  string
}

func (t *Title) Name() string { return "title" }
func (t *Title) ServiceName() string {
	return core.InterfacePrefix + t.Name()
}

func (t *Title) InterfaceName() string {
	return core.InterfacePrefix + t.Name() + ".Receiver"
}

func (t *Title) ServicePath() dbus.ObjectPath {
	return dbus.ObjectPath(path.Join(core.BaseServiceObjectPath, t.Name()))
}

func (t *Title) ServiceRun(conn *dbus.Conn, msg any) {
	t.CurrentName = msg.(string)
	data := fmt.Sprintf(`{ "name": "%s" }`, t.CurrentName)
	err := conn.Emit(t.ServicePath(), t.InterfaceName()+"."+t.Name(), data)
	if err != nil {
		log.Info(err)
	}
}

func (*Title) Synopsis() string { return "print hyprland current active client title as json" }
func (*Title) Usage() string {
	return `title
get hyprland current active client title as json
`
}

// SetFlags adds the check flags to the specified set.
func (*Title) SetFlags(f *flag.FlagSet) {
}

// Execute executes the check command.
func (m *Title) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
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
			message := signal.Body[0].(string)
			fmt.Println(message)
		}
	}
	return subcommands.ExitSuccess
}

func (t *Title) Update() *dbus.Error {
	t.updateSignal <- t.CurrentName
	return nil
}
