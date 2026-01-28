package forge

import (
	"git.sr.ht/~hannes/clonya/common"
	"git.sr.ht/~hannes/clonya/forge/github"
)

type Forge interface {
	Search(criteria common.SearchCriteria) ([]common.Repository, error)
	LatestCommitHash(id string) (string, error)
	Checkout(path string, repo common.Repository, full bool) error
	RepositoryDirname(repo common.Repository, full bool) string
}

func CreateClient(ty common.Forge) Forge {
	switch ty {
	case common.ForgeGithub:
		client := github.NewClient()
		return &client
	default:
		panic("unknown forge type")
	}
}
