package main

import (
	"context"
	"flag"
	"git.sr.ht/~hannes/clonya/forge"
	"github.com/google/subcommands"
	"log"
)

type updateCmd struct {
	dbPath string
}

func (*updateCmd) Name() string     { return "update" }
func (*updateCmd) Synopsis() string { return "Update the commit hashes in a repository database." }
func (*updateCmd) Usage() string {
	return `update: Update the commit hashes in a repository database.
`
}
func (u *updateCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&u.dbPath, "db", "clonya.se", "the path to the database")
}

func (u *updateCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	db, err := ReadDatabase(u.dbPath)
	if err != nil {
		log.Printf("unable to read database: %v\n", err)
		return subcommands.ExitFailure
	}

	exitCode := subcommands.ExitSuccess
	forgeClient := forge.CreateClient(db.searchCriteria.Forge)
	for i := range db.repositories {
		log.Println("updating commit hash for", db.repositories[i].Id)
		hash, err := forgeClient.LatestCommitHash(db.repositories[i].Id)
		if err != nil {
			log.Printf("warning: unable to update commit hash for %s: %v\n", db.repositories[i].Id, err)
			exitCode = subcommands.ExitFailure
			continue
		}
		db.repositories[i].CommitHash = hash
	}
	err = WriteDatabase(db, u.dbPath)
	if err != nil {
		log.Printf("unable to write database: %v\n", err)
		return subcommands.ExitFailure
	}
	return exitCode
}
