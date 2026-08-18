package main

import (
	"context"
	"flag"
	"git.sr.ht/~hannes/clonya/common"
	"git.sr.ht/~hannes/clonya/forge"
	"github.com/google/subcommands"
	"log"
	"time"
)

type initCmd struct {
	full           bool
	searchCriteria common.SearchCriteria
	dbPath         string
	maxCreateStr   string
	maxPushStr     string
	minCreateStr   string
	minPushStr     string
}

func (*initCmd) Name() string     { return "init" }
func (*initCmd) Synopsis() string { return "Initialize a repository database." }
func (*initCmd) Usage() string {
	return `init: Initialize a repository database.
`
}
func (i *initCmd) SetFlags(f *flag.FlagSet) {
	past := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.DateOnly)
	future := time.Now().Add(time.Duration(24 * time.Hour)).Format(time.DateOnly)

	f.StringVar(&i.dbPath, "db", "clonya.se", "the path to write the database to")
	f.StringVar((*string)(&i.searchCriteria.Forge), "forge", "github", "the forge to search")
	f.BoolVar(&i.searchCriteria.AllowForks, "forks", false, "allow forks")
	f.StringVar(&i.searchCriteria.Language, "lang", "go", "the main programming language of the repository")
	f.IntVar(&i.searchCriteria.Limit, "limit", 100, "the maximum number of repositories")
	f.StringVar(&i.maxCreateStr, "maxcreate", future, "the maximum creation date of the repository")
	f.StringVar(&i.maxPushStr, "maxpush", future, "the maximum push date of the repository")
	f.IntVar(&i.searchCriteria.MaxStars, "maxstars", -1, "the maximum number of stars")
	f.IntVar(&i.searchCriteria.MinCommits, "mincommits", 0, "the minimum number of commits")
	f.StringVar(&i.minCreateStr, "mincreate", past, "the minimum creation date of the repository")
	f.StringVar(&i.minPushStr, "minpush", past, "the minimum push date of the repository")
	f.IntVar(&i.searchCriteria.MinStars, "minstars", 0, "the minimum number of stars")
	f.BoolVar(&i.searchCriteria.AllowArchived, "no-archived", false, "disallow archived repositories")
	f.BoolVar(&i.full, "full", false, "clone the entire git repository in the checkout step instead of just a single commit")
}

func (i *initCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	date, err := time.Parse(time.DateOnly, i.minCreateStr)
	if err != nil {
		log.Printf("invalid date %s: %v\n", i.minCreateStr, err)
		return subcommands.ExitFailure
	}
	i.searchCriteria.MinCreateDate = date
	date, err = time.Parse(time.DateOnly, i.minPushStr)
	if err != nil {
		log.Printf("invalid date %s: %v\n", i.minPushStr, err)
		return subcommands.ExitFailure
	}
	i.searchCriteria.MinPushDate = date
	date, err = time.Parse(time.DateOnly, i.maxCreateStr)
	if err != nil {
		log.Printf("invalid date %s: %v\n", i.maxCreateStr, err)
		return subcommands.ExitFailure
	}
	i.searchCriteria.MaxCreateDate = date
	date, err = time.Parse(time.DateOnly, i.maxPushStr)
	if err != nil {
		log.Printf("invalid date %s: %v\n", i.maxPushStr, err)
		return subcommands.ExitFailure
	}
	i.searchCriteria.MaxPushDate = date
	i.searchCriteria.AllowArchived = !i.searchCriteria.AllowArchived

	forgeClient := forge.CreateClient(i.searchCriteria.Forge)
	repos, err := forgeClient.Search(i.searchCriteria)
	if err != nil {
		log.Println("unable to search for repositories:", err)
		return subcommands.ExitFailure
	}
	err = WriteDatabase(database{searchCriteria: i.searchCriteria, repositories: repos}, i.dbPath)
	if err != nil {
		log.Printf("unable to write database: %v\n", err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
