package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/go-co-op/gocron"
)

var globalScheduler *gocron.Scheduler

// GetGlobalScheduler returns the global scheduler, initializing it if needed.
func GetGlobalScheduler() *gocron.Scheduler {
	if globalScheduler == nil {
		globalScheduler = gocron.NewScheduler(time.UTC)
		globalScheduler.StartAsync()
	}
	return globalScheduler
}

func NotifySend(title, msg, iconPath string) {
	cmd := exec.Command("notify-send", title, msg, "-i", iconPath, "-c \"System\"", "-a \"System Notification\"")
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

var (
	posePath  string
	eyePath   string
	waterPath string
)

func init() {
	home, _ := os.UserHomeDir()

	posePath = filepath.Join(home, ".dotfiles/.config/eww/images/heart.png")
	eyePath = filepath.Join(home, ".dotfiles/.config/eww/images/eye.png")
	waterPath = filepath.Join(home, ".dotfiles/.config/eww/images/water.png")
}

func NotiPose() {
	NotifySend("Health", "Fix Your Posture!", posePath)
}

func NotiEye() {
	NotifySend("Health", "Look after your eyes!", eyePath)
}

func NotiWater() {
	NotifySend("Health", "Drink Water!", waterPath)
}
