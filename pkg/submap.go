package core

import (
	"context"
	"fmt"
	"log"
)

var SubmapInterfaceName = "com.hypr.submapService"

type SubmapService struct {
}

func DbusSubmap(ctx context.Context, submap chan string) {
	service := SubmapService{}
	conn := GetDbusConnection()
	err := conn.Export(service, ServiceObjectPath, SubmapInterfaceName)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case currentMap := <-submap:
			log.Print("submap signal received")
			if currentMap == "" {
				currentMap = "Default"
			}
			data := fmt.Sprintf(`{ "submap": "%s" }`, currentMap)
			err := conn.Emit(ServiceObjectPath, SubmapInterfaceName+".submap", data)
			if err != nil {
				log.Print(err)
			}
		}
	}
}
