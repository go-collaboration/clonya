package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
)

type versionCmd struct{}

func (*versionCmd) Name() string     { return "version" }
func (*versionCmd) Synopsis() string { return "Show the version number of the program." }
func (*versionCmd) Usage() string {
	return `version:
	Show the version number of the program.`
}
func (*versionCmd) SetFlags(*flag.FlagSet) {}
func (*versionCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	fmt.Println("clonya", Version)
	return subcommands.ExitSuccess
}
