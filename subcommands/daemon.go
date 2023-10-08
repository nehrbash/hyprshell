package subcommands

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	core "github.com/nehrbash/hyprshell/pkg"

	"github.com/google/subcommands"
)

type Daemon struct{}

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
		QuoteSignal:     make(chan string),
		TitleSignal:     make(chan string),
	}
	go hyprManager.HyprListen(ctx)
	go hyprManager.HyprClientListen(ctx)

	dock := &Dock{
		apps:         core.NewApps(),
		updateSignal: hyprManager.DockSignal,
		updateTitle:  hyprManager.TitleSignal,
	}
	submap := &Submap{}

	monitor := &Monitor{
		updateSignal: hyprManager.MonitorSignal,
	}

	workspaces := &Workspaces{
		workspaces:   core.NewWorkspaces(),
		updateSignal: hyprManager.WorkspaceSignal,
	}
	weather := &Weather{
		updateSignal: hyprManager.WeatherSignal,
		WeatherData:  "test",
	}
	title := &Title{
		updateSignal: hyprManager.TitleSignal,
		CurrentName:  "Keep up The Great Work!", // initialize just so it's never blank
	}

	quote := &Quote{QuoteSignal: hyprManager.QuoteSignal}

	go core.ServeCommand(ctx, dock, hyprManager.DockSignal)
	go core.ServeCommand(ctx, submap, hyprManager.SubmapSignal)
	go core.ServeCommand(ctx, monitor, hyprManager.MonitorSignal)
	go core.ServeCommand(ctx, workspaces, hyprManager.WorkspaceSignal)
	go core.ServeCommand(ctx, weather, hyprManager.WeatherSignal)
	go core.ServeCommand(ctx, quote, hyprManager.QuoteSignal)
	go core.ServeCommand(ctx, title, hyprManager.TitleSignal)

	cron := core.GetGlobalScheduler()
	cron.Every("25m").Do(func() {
		weather.Update()
	})

	cron.Every("1m").Do(func() {
		quote.Update()
	})
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
