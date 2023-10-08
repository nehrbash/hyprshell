package subcommands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/joshsziegler/zgo/pkg/log"

	core "github.com/nehrbash/hyprshell/pkg"

	"github.com/godbus/dbus/v5"
	"github.com/google/subcommands"
)

type Dock struct {
	// flags
	save         bool
	apps         core.Apps
	updateSignal chan string
	updateTitle  chan string
}

func (*Dock) Name() string { return "dock" }
func (d *Dock) ServiceName() string {
	return core.InterfacePrefix + d.Name()
}

func (d *Dock) InterfaceName() string {
	return core.InterfacePrefix + d.Name() + ".Receiver"
}

func (d *Dock) ServicePath() dbus.ObjectPath {
	return dbus.ObjectPath(path.Join(core.BaseServiceObjectPath, d.Name()))
}

func (d *Dock) ServiceRun(conn *dbus.Conn, msg any) {
	applist := d.apps.Update(core.GetClients())
	// send title data
	d.updateTitle <- core.FocusedClient().Title
	conn.Emit(d.ServicePath(), d.InterfaceName()+"."+d.Name(), applist.String())
}

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
func (d *Dock) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	conn, err := dbus.SessionBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		return subcommands.ExitSuccess
	}

	if d.save {
		obj := conn.Object(d.ServiceName(), d.ServicePath())
		err = obj.Call("DoAction", 0, "save").Store()
		if err != nil {
			log.Info(err)
		}

	} else { // default stream data

		conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
			fmt.Sprintf("type='signal',sender='%s',interface='%s',path='%s',member='%s'",
				d.ServiceName(), d.InterfaceName(), d.ServicePath(), d.Name()))
		signalChan := make(chan *dbus.Signal)
		conn.Signal(signalChan)

		// request first quote
		obj := conn.Object(d.ServiceName(), d.ServicePath())
		err = obj.Call("Update", 0).Store()
		if err != nil {
			log.Info(err)
		}

		for signal := range signalChan {
			if signal.Name == d.InterfaceName()+"."+d.Name() {
				fmt.Println(signal.Body[0].(string))
			}
		}
	}
	return subcommands.ExitSuccess
}

func (d *Dock) Update() *dbus.Error {
	d.updateSignal <- "request update bitch"
	return nil
}

func (d *Dock) DoAction(action string) *dbus.Error {
	switch action {
	case "save":
		data, err := core.FavoritsMarshalJSON(d.apps.Update(core.GetClients()))
		if err != nil {
			log.Error(err)
			return nil
		}
		err = core.SaveFavorites(data)
		if err != nil {
			log.Error(err)
			return nil
		}
	default:
		fmt.Printf("Msg: %s \n", action)
	}
	return nil
}
