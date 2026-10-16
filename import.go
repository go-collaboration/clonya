package main

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	"git.sr.ht/~hannes/clonya/common"
	"github.com/google/subcommands"
)

type searchResults struct {
	Parameters map[string]json.RawMessage `json:"parameters"`
	Items      []searchResultItem         `json:"items"`
}

type searchResultItem struct {
	Name          string `json:"name"`
	IsArchived    bool   `json:"isArchived"`
	Stargazers    int    `json:"stargazers"`
	LastCommitSHA string `json:"lastCommitSHA"`
}

type importCmd struct {
	full       bool
	dbPath     string
	forge      string
	jsonPath   string
	limit      int
	noArchived bool
}

func (*importCmd) Name() string     { return "import" }
func (*importCmd) Synopsis() string { return "Convert JSON search results into a repository database." }
func (*importCmd) Usage() string {
	return `import: Convert JSON search results into a repository database.
`
}
func (i *importCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&i.dbPath, "db", "clonya.se", "the path to write the database to")
	f.StringVar(&i.forge, "forge", "github", "the forge the results originate from")
	f.StringVar(&i.jsonPath, "json", "results.json", "the path to the JSON search results")
	f.IntVar(&i.limit, "limit", 0, "the maximum number of repositories to import (0 imports all of them)")
	f.BoolVar(&i.noArchived, "no-archived", false, "disallow archived repositories")
	f.BoolVar(&i.full, "full", false, "clone the entire git repository in the checkout step instead of just a single commit")
}

func (i *importCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	results, err := readSearchResults(i.jsonPath)
	if err != nil {
		log.Printf("unable to read search results: %v\n", err)
		return subcommands.ExitFailure
	}

	past := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	future := time.Now().Add(time.Duration(24 * time.Hour))
	searchCriteria := common.SearchCriteria{
		Forge:         common.Forge(i.forge),
		MinCreateDate: past,
		MaxCreateDate: future,
		MinPushDate:   past,
		MaxPushDate:   future,
		MaxStars:      -1,
		AllowArchived: !i.noArchived,
	}
	err = parseSearchParameters(&searchCriteria, results.Parameters)
	if err != nil {
		log.Printf("unable to read search parameters: %v\n", err)
		return subcommands.ExitFailure
	}

	slices.SortStableFunc(results.Items, func(a, b searchResultItem) int {
		return cmp.Compare(b.Stargazers, a.Stargazers)
	})

	repos := make([]common.Repository, 0, len(results.Items))
	for _, item := range results.Items {
		if i.limit > 0 && len(repos) >= i.limit {
			break
		}
		if item.Name == "" || item.LastCommitSHA == "" {
			log.Printf("warning: skipping repository %s without a commit hash\n", item.Name)
			continue
		}
		if item.IsArchived && !searchCriteria.AllowArchived {
			continue
		}
		repos = append(repos, common.Repository{Id: item.Name, CommitHash: item.LastCommitSHA})
	}
	searchCriteria.Limit = len(repos)

	err = WriteDatabase(database{searchCriteria: searchCriteria, full: i.full, repositories: repos}, i.dbPath)
	if err != nil {
		log.Printf("unable to write database: %v\n", err)
		return subcommands.ExitFailure
	}
	log.Println("imported", len(repos), "repositories")
	return subcommands.ExitSuccess
}

func readSearchResults(path string) (searchResults, error) {
	results := searchResults{}

	input, err := os.Open(path)
	if err != nil {
		return results, err
	}
	defer input.Close()

	err = json.NewDecoder(input).Decode(&results)
	if err != nil {
		return results, err
	}

	return results, nil
}

func parseSearchParameters(criteria *common.SearchCriteria, parameters map[string]json.RawMessage) error {
	for name, value := range parameters {
		if string(value) == "null" {
			continue
		}
		err := error(nil)
		switch name {
		case "language":
			err = json.Unmarshal(value, &criteria.Language)
		case "commitsMin":
			err = json.Unmarshal(value, &criteria.MinCommits)
		case "starsMin":
			err = json.Unmarshal(value, &criteria.MinStars)
		case "starsMax":
			err = json.Unmarshal(value, &criteria.MaxStars)
		case "createdMin":
			criteria.MinCreateDate, err = parseTimestamp(value)
		case "createdMax":
			criteria.MaxCreateDate, err = parseTimestamp(value)
		case "committedMin":
			criteria.MinPushDate, err = parseTimestamp(value)
		case "committedMax":
			criteria.MaxPushDate, err = parseTimestamp(value)
		case "excludeForks":
			excludeForks := false
			err = json.Unmarshal(value, &excludeForks)
			criteria.AllowForks = !excludeForks
		case "onlyForks":
			onlyForks := false
			err = json.Unmarshal(value, &onlyForks)
			if onlyForks {
				criteria.AllowForks = true
			}
		default:
			if !isUnsetParameter(value) {
				log.Printf("warning: ignoring unsupported search parameter %s\n", name)
			}
		}
		if err != nil {
			return fmt.Errorf("unable to read search parameter %s: %v", name, err)
		}
	}
	return nil
}

func isUnsetParameter(value json.RawMessage) bool {
	switch string(value) {
	case "false", "0", `""`, "[]":
		return true
	}
	return false
}

func parseTimestamp(value json.RawMessage) (time.Time, error) {
	milliseconds := int64(0)
	err := json.Unmarshal(value, &milliseconds)
	if err == nil {
		return time.UnixMilli(milliseconds).UTC(), nil
	}

	text := ""
	err = json.Unmarshal(value, &text)
	if err != nil {
		return time.Time{}, errors.New("expected a timestamp")
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", time.DateOnly} {
		timestamp, err := time.Parse(layout, text)
		if err == nil {
			return timestamp, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid timestamp %s", text)
}
