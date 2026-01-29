package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"slices"

	"git.sr.ht/~hannes/clonya/forge"
	"github.com/google/subcommands"
)

type checkoutCmd struct {
	clean    bool
	dbPath   string
	repoPath string
}

func (*checkoutCmd) Name() string     { return "checkout" }
func (*checkoutCmd) Synopsis() string { return "Checkout the repositories listed in a database." }
func (*checkoutCmd) Usage() string {
	return `checkout: Checkout the repositories listed in a database.
`
}
func (c *checkoutCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&c.clean, "clean", false, "whether to cleanup old checkouts")
	f.StringVar(&c.dbPath, "db", "clonya.se", "the path to the database")
	f.StringVar(&c.repoPath, "out", "repos", "the path to the output directory")
}

func (c *checkoutCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	db, err := ReadDatabase(c.dbPath)
	if err != nil {
		log.Printf("unable to read database: %v\n", err)
		return subcommands.ExitFailure
	}

	if _, err := os.Stat(c.repoPath); os.IsNotExist(err) {
		err = os.MkdirAll(c.repoPath, 0755)
		if err != nil {
			log.Printf("unable to create output directory: %v\n", err)
			return subcommands.ExitFailure
		}
	}

	exitCode := subcommands.ExitSuccess
	forgeClient := forge.CreateClient(db.searchCriteria.Forge)

	if c.clean {
		expectedNames := make([]string, 0, len(db.repositories))
		for _, repo := range db.repositories {
			expectedNames = append(expectedNames, forgeClient.RepositoryDirname(repo, db.full))
		}

		storedCheckouts, err := os.ReadDir(c.repoPath)
		if err != nil {
			log.Printf("unable to read directory entries in %s: %v\n", c.repoPath, err)
			return subcommands.ExitFailure
		}
		for _, dir := range storedCheckouts {
			if !dir.Type().IsDir() {
				continue
			}
			if !slices.Contains(expectedNames, dir.Name()) {
				log.Println("deleting outdated repository checkout", dir.Name())
				os.RemoveAll(filepath.Join(c.repoPath, dir.Name()))
			}
		}
	}

	for _, repo := range db.repositories {
		log.Println("checking out", repo.Id)
		err := forgeClient.Checkout(c.repoPath, repo, db.full)
		if err != nil {
			log.Printf("warning: unable to checkout %s: %v\n", repo.Id, err)
			exitCode = subcommands.ExitFailure
			continue
		}
	}

	return exitCode
}
