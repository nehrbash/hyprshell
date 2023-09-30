package subcommands

import (
	"context"
	"flag"
	"fmt"

	"github.com/joshsziegler/zgo/pkg/log"
	core "github.com/nehrbash/hyprshell/pkg"
	"github.com/nehrbash/hyprshell/pkg/weather"
	"os"
	"path"

	"github.com/godbus/dbus/v5"
	"github.com/google/subcommands"
)

type Weather struct {
	updateSignal chan string
	WeatherData  string
}

func (*Weather) Name() string { return "weather" }

func (w *Weather) ServiceName() string {
	return core.InterfacePrefix + w.Name()
}

func (w *Weather) InterfaceName() string {
	return core.InterfacePrefix + w.Name() + ".Receiver"
}
func (w *Weather) ServicePath() dbus.ObjectPath {
	return dbus.ObjectPath(path.Join(core.BaseServiceObjectPath, w.Name()))
}
func (w *Weather) ServiceRun(conn *dbus.Conn, msg any) {
	log.Info("weather update")
	err := conn.Emit(w.ServicePath(), w.InterfaceName()+"."+w.Name(), w.WeatherData)
	if err != nil {
		log.Info(err)
	}
}

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

	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		fmt.Sprintf("type='signal',sender='%s',interface='%s',path='%s',member='%s'",
			w.ServiceName(), w.InterfaceName(), w.ServicePath(), w.Name()))
	signalChan := make(chan *dbus.Signal)
	conn.Signal(signalChan)

	// request first quote
	obj := conn.Object(w.ServiceName(), w.ServicePath())
	err = obj.Call("GetWeather", 0).Store()
	if err != nil {
		log.Info(err)
	}

	// handle
	for signal := range signalChan {
		if signal.Name == w.InterfaceName()+"."+w.Name() {
			fmt.Println(signal.Body[0].(string))
		}
	}

	return subcommands.ExitSuccess
}

func (w *Weather) Update() {
	RawWeather, err := weather.GetWeatherData()
	if err != nil {
		log.Infof("Can not get weather data  %s", err.Error())
		return
	}
	ewwData := weather.NewEwwVariables(RawWeather)
	w.WeatherData = ewwData.String()
	w.updateSignal <- "sending update"
}

func (w *Weather) GetWeather() (err *dbus.Error) {
	w.updateSignal <- "update plz"
	return nil
}
