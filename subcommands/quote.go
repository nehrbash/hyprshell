package subcommands

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	core "github.com/nehrbash/hyprshell/pkg"
	"github.com/nehrbash/hyprshell/pkg/quote"

	"github.com/godbus/dbus/v5"
	"github.com/google/subcommands"
)

// Submap is a comannd manager struct
type Quote struct {
	Author      string
	Quote       string
	QuoteSignal chan string
}

func (*Quote) Name() string { return "quote" }
func (q *Quote) ServiceName() string {
	return core.InterfacePrefix + q.Name()
}

func (q *Quote) InterfaceName() string {
	return core.InterfacePrefix + q.Name() + ".Receiver"
}
func (q *Quote) ServicePath() dbus.ObjectPath {
	return dbus.ObjectPath(path.Join(core.BaseServiceObjectPath, q.Name()))
}
func (q *Quote) ServiceRun(conn *dbus.Conn, msg any) {
	err := conn.Emit(q.ServicePath(), q.InterfaceName()+"."+q.Name(), msg)
	if err != nil {
		log.Print(err)
	}
}

func (*Quote) Synopsis() string { return "json quote with author" }
func (*Quote) Usage() string {
	return `quote
json Quote and Author
`
}

// SetFlags adds the check flags to the specified set.
func (q *Quote) SetFlags(f *flag.FlagSet) {
}

// Execute executes the check command.
func (q *Quote) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	conn, err := dbus.SessionBus()

	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		return subcommands.ExitSuccess
	}

	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		fmt.Sprintf("type='signal',sender='%s',interface='%s',path='%s',member='%s'",
			q.ServiceName(), q.InterfaceName(), q.ServicePath(), q.Name()))
	signalChan := make(chan *dbus.Signal)
	conn.Signal(signalChan)

	// request first quote
	obj := conn.Object(q.ServiceName(), q.ServicePath())
	err = obj.Call("GetQuote", 0, "get").Store()
	if err != nil {
		log.Print(err)
	}

	// listen
	for signal := range signalChan {
		if signal.Name == q.InterfaceName()+"."+q.Name() {
			message := signal.Body[0].(string)
			fmt.Println(message)
		}
	}
	return subcommands.ExitSuccess
}

func (q *Quote) Update() {
	var err error
	q.Quote, q.Author, err = quote.GetRandomQuote()
	if err != nil {
		log.Print(err)
	}
	q.QuoteSignal <- fmt.Sprintf(`{ "quote": "%s", "author": "%s" }`, q.Quote, q.Author)
}

func (q *Quote) GetQuote(msg string) (err *dbus.Error) {
	q.Update()
	return nil
}
