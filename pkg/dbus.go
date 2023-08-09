package core

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/godbus/dbus/v5"
)

var (
	globalDbusConn    *dbus.Conn
	globalDbusConnMux sync.Mutex
)

const (
	ServiceName       = "com.hypr.HelperService"
	ServiceObjectPath = "/com/hypr/HelperService"
	ServiceInterface  = "com.hypr.HelperService"
)

func GetDbusConnection() *dbus.Conn {
	var err error
	globalDbusConnMux.Lock()
	defer globalDbusConnMux.Unlock()
	if globalDbusConn == nil {
		globalDbusConn, err = dbus.SessionBus()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
			os.Exit(1)
		}
		reply, err := globalDbusConn.RequestName(ServiceName, dbus.NameFlagDoNotQueue|dbus.NameFlagReplaceExisting)
		if err != nil {
			log.Fatal("Failed to request name:", err)
		}

		if reply != dbus.RequestNameReplyPrimaryOwner {
			log.Fatal("Name already taken")
		}

	}
	return globalDbusConn
}

func GetDbusListener[T any](signalName string) chan *T {
	returnChan := make(chan *T)
	if globalDbusConn == nil {
		globalDbusConn, err := dbus.SessionBus()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
			os.Exit(1)
		}

		globalDbusConn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
			fmt.Sprintf("type='signal',sender='%s',interface='%s',path='%s',member='%s'",
				ServiceName, ServiceInterface, ServiceObjectPath, signalName))

		signalChan := make(chan *dbus.Signal, 10)

		globalDbusConn.Signal(signalChan)
		for signal := range signalChan {
			if signal.Name == ServiceInterface+"."+signalName {
				returnChan <- signal.Body[0].(*T)
			}
		}

	}
	return returnChan
}
