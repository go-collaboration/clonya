package common

import "time"

type Forge string

const (
	ForgeGithub Forge = "github"
)

type SearchCriteria struct {
	Limit         int
	Forge         Forge
	Language      string
	MinCreateDate time.Time
	MaxCreateDate time.Time
	MinPushDate   time.Time
	MaxPushDate   time.Time
	AllowForks    bool
	MinStars      int
	MaxStars      int
	MinCommits    int
	AllowArchived bool
}

type Repository struct {
	Id         string
	CommitHash string
}
