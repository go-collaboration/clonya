package main

import (
	"context"
	"flag"
	"github.com/google/subcommands"
	"os"
)

const Version = "0.1.2"

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(&initCmd{}, "")
	subcommands.Register(&importCmd{}, "")
	subcommands.Register(&updateCmd{}, "")
	subcommands.Register(&upgradeCmd{}, "")
	subcommands.Register(&checkoutCmd{}, "")
	subcommands.Register(&versionCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
