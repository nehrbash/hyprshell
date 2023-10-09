package core

import (
	"time"

	"github.com/gen2brain/beeep"
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

func NotifySend(title, message, iconPath string) {
	beeep.Notify(title, message, iconPath)
}

// TODO embed images
func NotiPose() {
	NotifySend("Health", "Fix Your Posture!", "/home/nehrbash/.dotfiles/.config/eww/images/heart.png")
}

func NotiEye() {
	NotifySend("Health", "Loook after your eyes!", "/home/nehrbash/.dotfiles/.config/eww/images/heart.png")
}

func NotiWater() {
	NotifySend("Health", "Drink Water!", "/home/nehrbash/.dotfiles/.config/eww/images/heart.png")
}
