package subcommands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/subcommands"
)

const (
	// Constants for reading battery information from the system
	batteryDir       = "/sys/class/power_supply/BAT0"
	statusPath       = batteryDir + "/status"
	currentNowPath   = batteryDir + "/current_now"
	capacityPath     = batteryDir + "/capacity"
	statusCharging   = "Charging"
	statusFull       = "full"
	colorLowBattery  = "#f38ba8"
	colorHighBattery = "#a6e3a1"
)

// Define slice of batteryIcons
var (
	batteryOff = "󱉝"
	// low 10, 20,30 ,40,50,60,70,80,90,full
	batteryIcons         = []string{"󱊡", "󰁺", "󰁻", "󰁼", "󰁽", "󰁾", "󰁿", "󰂀", "󰂁", "󰂂", "󰁹"}
	BatteryChargingicons = []string{"󱊤", "󰢜", "󰂆", "󰂇", "󰂈", "󰢝", "󰂉", "󰢞", "󰂊", "󰂋", "󰂅"}
)

// Define slice of icons
// BatteryStatus represents the battery information
type BatteryStatus struct {
	Enabled    bool    `json:"enabled"`
	Icon       string  `json:"icon"`
	Percentage int     `json:"percentage"`
	Wattage    float64 `json:"wattage"`
	Status     string  `json:"status"`
	Color      string  `json:"color"`
}

// readFile reads the content of a file as a string
func readFile(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// isEnabled checks if the battery directory exists
func isEnabled() bool {
	_, err := os.Stat(batteryDir)
	return !os.IsNotExist(err)
}

// getState gets the charging state of the battery
func getState() (string, error) {
	return readFile(statusPath)
}

// getCurrentNow gets the current charge rate of the battery
func getCurrentNow() (int, error) {
	rateStr, err := readFile(currentNowPath)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(rateStr)
}

// getCapacity gets the current capacity level of the battery
func getCapacity() (int, error) {
	capacityStr, err := readFile(capacityPath)
	if err != nil {
		return 0, err
	}
	val, err := strconv.Atoi(capacityStr)
	if val > 100 || err != nil {
		return 100, err
	}
	return val, nil
}

type Battery struct{}

func (*Battery) Name() string     { return "battery" }
func (*Battery) Synopsis() string { return "current battery data" }
func (*Battery) Usage() string {
	return `title
get current app title as json data: Stream
`
}

// SetFlags adds the check flags to the specified set.
func (m *Battery) SetFlags(f *flag.FlagSet) {
}

func (b *BatteryStatus) String() string {
	b.Enabled = isEnabled()

	if !b.Enabled {
		jsonOutput, _ := json.Marshal(b)
		return string(jsonOutput)

	}

	var err error
	b.Status, err = getState()
	if err != nil {
		log.Fatal(err)
	}

	rate, err := getCurrentNow()
	if err != nil {
		log.Fatal(err)
	}

	capacity, err := getCapacity() // in percentage
	if err != nil {
		log.Fatal(err)
	}

	level := int(math.Round(float64(capacity)/10.0)) * 10
	if capacity >= 95 {
		level += 1 // full icon
	}
	if rate > 0 { // charging
		b.Icon = BatteryChargingicons[level]
	} else {
		b.Icon = batteryIcons[level]
	}

	b.Percentage = capacity
	b.Wattage = float64(rate) / 1000000.0

	if capacity <= 20 {
		b.Color = colorLowBattery
	} else {
		b.Color = colorHighBattery
	}

	jsonOutput, _ := json.Marshal(b)
	return string(jsonOutput)
}

// Execute executes the check command.
func (m *Battery) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	// default output
	status := BatteryStatus{
		Icon:   batteryOff,
		Color:  colorHighBattery,
		Status: "N/A",
	}
	fmt.Println(status.String())

	for {
		select {
		case <-ticker.C:
			fmt.Println(status.String())
		case <-ctx.Done():
			return subcommands.ExitSuccess

		}
	}
}
