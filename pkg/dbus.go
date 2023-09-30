package core

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/godbus/dbus/v5"
)

var (
	globalDbusConn *dbus.Conn
)

const (
	InterfacePrefix       = "com.hypr."
	BaseServiceObjectPath = "/com/hypr"

	ServiceInterface = "com.hypr.HelperService"
)

type DbusCommand interface {
	ServicePath() dbus.ObjectPath
	ServiceName() string
	InterfaceName() string
	ServiceRun(conn *dbus.Conn, msg any)
}

func ServeCommand[T any](ctx context.Context, service DbusCommand, signal chan T) {
	conn := GetDbusConnection(service.ServiceName())
	err := conn.Export(service, service.ServicePath(), service.InterfaceName())
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-signal:
			service.ServiceRun(conn, msg)
		}
	}
}

func GetDbusConnection(serviceName string) *dbus.Conn {
	var err error
	if globalDbusConn == nil {
		globalDbusConn, err = dbus.SessionBus()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
			os.Exit(1)
		}
		reply, err := globalDbusConn.RequestName(serviceName, dbus.NameFlagDoNotQueue|dbus.NameFlagReplaceExisting)
		if err != nil {
			log.Fatal("Failed to request name:", err)
		}

		if reply != dbus.RequestNameReplyPrimaryOwner {
			log.Fatal("Name already taken")
		}

	}
	return globalDbusConn
}

func GetDbusListener[T any](serviceName, servicePath, signalName string) chan T {
	returnChan := make(chan T)
	if globalDbusConn == nil {
		globalDbusConn, err := dbus.SessionBus()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
			os.Exit(1)
		}

		globalDbusConn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
			fmt.Sprintf("type='signal',sender='%s',interface='%s',path='%s',member='%s'",
				serviceName, ServiceInterface, servicePath, signalName))

		signalChan := make(chan *dbus.Signal, 10)

		globalDbusConn.Signal(signalChan)
		for signal := range signalChan {
			if signal.Name == ServiceInterface+"."+signalName {
				returnChan <- signal.Body[0].(T)
			}
		}

	}
	return returnChan
}
