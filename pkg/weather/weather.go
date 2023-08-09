package weather

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type EwwVariables struct {
	Location    string  `json:"location"`
	Feel        string  `json:"feelTemp"`
	Current     string  `json:"currentTemp"`
	CurrentIcon string  `json:"currentIcon"`
	CurrentDesc string  `json:"currentDesc"`
	MoonLable   string  `json:"moonLabel"`
	Moons       [8]Moon `json:"moons"`
	Hours       []Hour  `json:"hours"`
}

type Hour struct {
	Temp string `json:"temp"`
	Icon string `json:"icon"`
	Time string `json:"time"`
}

type Moon struct {
	Class string `json:"class"`
	Value string `json:"value"`
}

// Only gives 3 day forcast
func NewEwwVariables(w WeatherForcast) EwwVariables {
	return EwwVariables{
		Location:    w.NearestArea[0].AreaName[0].Value,
		Feel:        w.CurrentCondition[0].FeelsLikeF,
		Current:     w.CurrentCondition[0].TempF,
		CurrentIcon: Icon(w.CurrentCondition[0].WeatherCode),
		CurrentDesc: w.CurrentCondition[0].WeatherDesc[0].Value,
		MoonLable:   w.Weather[0].Astronomy[0].MoonPhase,
		Moons:       Moons(w.Weather[0].Astronomy[0].MoonPhase),
		Hours:       w.Hours(),
	}
}

func (w *WeatherForcast) Hours() []Hour {
	days := len(w.Weather[0].Hourly)
	t := time.Now()
	// Round to the nearest hour
	t = t.Add(30 * time.Minute)
	t = t.Round(time.Hour)
	var hours = make([]Hour, days, days)
	for i, h := range w.Weather[0].Hourly {
		hours[i].Temp = h.TempF
		hours[i].Icon = Icon(h.WeatherCode)
		// docs don't fucking say what the time of the hours are so lets just make them up
		hours[i].Time = t.Add(time.Hour * time.Duration(i)).Format("03:00")
	}
	// 8 is a lot drop the first and last one
	if days > 6 {
		hours = hours[1 : days-1]
	}
	return hours
}

func Moons(phase string) (moons [8]Moon) {
	cur := MoonIdx(phase)
	for i, m := range MOON_PHASES {
		if i == cur {
			moons[i] = Moon{
				Value: m,
				Class: "active-moon",
			}
		} else {
			moons[i] = Moon{
				Value: m,
			}
		}
	}
	return
}

func (e *EwwVariables) String() string {
	json, err := json.Marshal(e)
	if err != nil {
		log.Printf("Marshalling Weather Data: %s", err.Error())
		return ""
	}
	return string(json)
}

func GetWeatherData() (w WeatherForcast, err error) {
	resp, err := http.Get("http://wttr.in/?format=j1")
	if err != nil {
		return w, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body) // response body is []byte
	if err != nil {
		return w, err
	}

	var result WeatherForcast
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte tostruct
		return w, err
	}
	return result, nil
}
