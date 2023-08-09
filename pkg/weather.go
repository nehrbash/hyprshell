package core

import (
	"context"
	"log"

	"github.com/godbus/dbus/v5"
	"github.com/nehrbash/hyprshell/pkg/weather"
)

const ServiceInterfaceWeather = "com.hypr.WeatherService"

type WeatherService struct {
	weatherEvent chan string
	WeatherData  string
}

func (w *WeatherService) update() {
	RawWeather, err := weather.GetWeatherData()
	if err != nil {
		log.Printf("Can not get weather data  %s", err.Error())
		return
	}
	ewwData := weather.NewEwwVariables(RawWeather)
	w.WeatherData = ewwData.String()
	w.weatherEvent <- w.WeatherData
}

func (w *WeatherService) GetWeather(msg string) (err *dbus.Error) {
	w.weatherEvent <- w.WeatherData
	return nil
}

func DbusWeather(ctx context.Context) {
	cron := GetGlobalScheduler()
	service := &WeatherService{weatherEvent: make(chan string, 2)}
	cron.Every("25m").Do(func() {
		service.update()
	})

	conn := GetDbusConnection()
	err := conn.Export(service, ServiceObjectPath, ServiceInterfaceWeather)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-service.weatherEvent:
			err := conn.Emit(ServiceObjectPath, ServiceInterfaceWeather+".Weather", data)
			if err != nil {
				log.Print(err)
			}
		}
	}
}
