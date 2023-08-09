package core

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/joshsziegler/zgo/pkg/log"
)

type DockService struct {
	apps Apps
}

func (d *DockService) DoAction(action string) *dbus.Error {
	switch action {
	case "save":
		data, err := FavoritsMarshalJSON(d.apps.Update(GetClients()))
		if err != nil {
			log.Error(err)
			return nil
		}
		err = saveFavorites(data)
		if err != nil {
			log.Error(err)
			return nil
		}
	default:
		fmt.Printf("Msg: %s \n", action)
	}
	return nil
}

func DbusDock(ctx context.Context, dockEvent chan string, signalName string) {
	conn := GetDbusConnection()
	service := &DockService{
		apps: NewApps(),
	}
	conn.Export(service, ServiceObjectPath, ServiceInterface)
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-dockEvent:
			switch event {
			default: // update

			}
			applist := service.apps.Update(GetClients())
			conn.Emit(ServiceObjectPath, ServiceInterface+"."+signalName, applist.String())
		}
	}
}
