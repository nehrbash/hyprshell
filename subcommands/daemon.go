package subcommands

import (
	"context"
	"flag"
	core "hyprshell/pkg"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/subcommands"
)

type Daemon struct {
}

func (*Daemon) Name() string     { return "daemon" }
func (*Daemon) Synopsis() string { return "start the desktop managing daemon" }
func (*Daemon) Usage() string {
	return `daemon
  Start the daemon
`
}

func (*Daemon) SetFlags(f *flag.FlagSet) {
	// TODO: Add config file, probably yaml
}

func (*Daemon) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	signalNotify := make(chan os.Signal, 1)
	signal.Notify(signalNotify, syscall.SIGINT, syscall.SIGTERM)

	hyprManager := core.HyprSignalManager{
		HyprEvent:       make(chan core.HyprSignal),
		DockSignal:      make(chan string),
		WeatherSignal:   make(chan string),
		SubmapSignal:    make(chan string),
		MonitorSignal:   make(chan string),
		WorkspaceSignal: make(chan string),
	}

	go hyprManager.HyprListen(ctx)
	go hyprManager.HyprClientListen(ctx)
	go core.DbusDock(ctx, hyprManager.DockSignal, "dock")
	go core.DbusSubmap(ctx, hyprManager.SubmapSignal)
	go core.DbusMontitor(ctx, hyprManager.MonitorSignal)
	go core.DbusWorkspaces(ctx, hyprManager.WorkspaceSignal)
	go core.DbusWeather(ctx)
	go core.DbusQuote(ctx)

	cron := core.GetGlobalScheduler()
	cron.Every("25m").Do(func() {
		time.Sleep(time.Duration(rand.Int63n(int64(5 * time.Minute))))
		core.NotiPose()
	})
	cron.Every("30m").Do(func() {
		time.Sleep(time.Duration(rand.Int63n(int64(5 * time.Minute))))
		core.NotiEye()
	})

	cron.Every("2h").Do(func() {
		time.Sleep(time.Duration(rand.Int63n(int64(5 * time.Minute))))
		core.NotiEye()
	})

	// block and wait for kill signal
	func() {
		for {
			s := <-signalNotify
			switch s {
			case syscall.SIGTERM, syscall.SIGINT:
				log.Printf("`%s` received, bye bye!", s)
				// save faviorits and whatever setting I add
				ctx.Done()
				return
			default:
				log.Print("Unknown signal")
			}
		}
	}()
	return subcommands.ExitSuccess
}
