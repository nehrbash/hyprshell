package main

import (
	"context"
	"flag"
	"os"

	cmd "github.com/nehrbash/hyprshell/subcommands"

	sc "github.com/google/subcommands"
)

var (
	Version   string // Latest Git tag (e.g. v1.0.1)
	BuildDate string // Date the binary was built
)

func init() {
	sc.Register(sc.HelpCommand(), "")
	sc.Register(sc.FlagsCommand(), "")
	sc.Register(sc.CommandsCommand(), "")
	sc.Register(&cmd.Daemon{}, "")
	sc.Register(&cmd.Weather{}, "")
	sc.Register(&cmd.Dock{}, "")
	sc.Register(&cmd.Submap{}, "")
	sc.Register(&cmd.Monitor{}, "")
	sc.Register(&cmd.Workspaces{}, "")
	sc.Register(&cmd.Quote{}, "")
	sc.Register(&cmd.Battery{}, "")
	flag.Parse()
}

func main() {
	ctx := context.Background()
	os.Exit(int(sc.Execute(ctx)))
}
