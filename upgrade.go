package main

import (
	"context"
	"flag"
	"git.sr.ht/~hannes/clonya/forge"
	"github.com/google/subcommands"
	"log"
)

type upgradeCmd struct {
	dbPath string
}

func (*upgradeCmd) Name() string     { return "upgrade" }
func (*upgradeCmd) Synopsis() string { return "Update the repository list in a database." }
func (*upgradeCmd) Usage() string {
	return `upgrade: Update the repository list in a database.
`
}
func (u *upgradeCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&u.dbPath, "db", "clonya.se", "the path to the database")
}

func (u *upgradeCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	db, err := ReadDatabase(u.dbPath)
	if err != nil {
		log.Printf("unable to read database: %v\n", err)
		return subcommands.ExitFailure
	}

	forgeClient := forge.CreateClient(db.searchCriteria.Forge)
	repos, err := forgeClient.Search(db.searchCriteria)
	if err != nil {
		log.Println("unable to search for repositories:", err)
		return subcommands.ExitFailure
	}
	db.repositories = repos

	err = WriteDatabase(db, u.dbPath)
	if err != nil {
		log.Printf("unable to write database: %v\n", err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
