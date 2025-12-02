package github

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"git.sr.ht/~hannes/clonya/common"
	"github.com/google/go-github/v76/github"
)

func NewClient() Client {
	githubClient := github.NewClient(nil)
	accessToken, found := os.LookupEnv("CLONYA_GITHUB_ACCESS_TOKEN")
	if found {
		githubClient = githubClient.WithAuthToken(accessToken)
	}
	return Client{client: githubClient}
}

type Client struct {
	client *github.Client
}

func (c *Client) Search(criteria common.SearchCriteria) ([]common.Repository, error) {
	repos := make([]common.Repository, 0, criteria.Limit)
	ctx := context.WithValue(context.Background(), github.SleepUntilPrimaryRateLimitResetWhenRateLimited, true)

	opt := &github.SearchOptions{ListOptions: github.ListOptions{PerPage: 100}, Sort: "stars"}

	// Construct search query
	searchQuery := "language:" + criteria.Language
	if criteria.MinCreateDate.After(time.Date(2008, 4, 1, 0, 0, 0, 0, time.UTC)) {
		searchQuery += " created:>" + criteria.MinCreateDate.Format(time.DateOnly)
	}
	if criteria.MaxCreateDate.Before(time.Now()) {
		searchQuery += " created:<" + criteria.MaxCreateDate.Format(time.DateOnly)
	}
	if criteria.MinPushDate.After(time.Date(2008, 4, 1, 0, 0, 0, 0, time.UTC)) {
		searchQuery += " pushed:>" + criteria.MinPushDate.Format(time.DateOnly)
	}
	if criteria.MaxPushDate.Before(time.Now()) {
		searchQuery += " pushed:<" + criteria.MaxPushDate.Format(time.DateOnly)
	}
	if criteria.MinStars > 0 {
		searchQuery += fmt.Sprintf(" stars:>%d", criteria.MinStars)
	}
	if criteria.MaxStars > 0 {
		searchQuery += fmt.Sprintf(" stars:<=%d", criteria.MaxStars)
	}
	if !criteria.AllowArchived {
		searchQuery += " archived:false"
	}
	if criteria.AllowForks {
		searchQuery += " fork:true"
	}
	log.Println("built search query:", searchQuery)

	for {
		log.Println("searching for next repositories")
		foundRepos, resp, err := c.client.Search.Repositories(ctx, searchQuery, opt)
		if isRateLimitError(err) {
			continue
		} else if err != nil {
			return repos, err
		}

		// Process results
		for _, repo := range (*foundRepos).Repositories {
			id := repo.GetFullName()
			log.Println("retrieving latest commit hash for", id)
			hash, err := c.LatestCommitHashForBranch(id, repo.GetDefaultBranch())
			if err != nil {
				log.Printf("warning: unable to find commit hash for %s, skipping: %v\n", id, err)
				continue
			}
			repos = append(repos, common.Repository{Id: id, CommitHash: hash})
			if len(repos) >= criteria.Limit {
				break
			}
		}

		if len(repos) >= criteria.Limit {
			break
		}

		// Get next page if available
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return repos, nil
}

func (c *Client) LatestCommitHash(id string) (string, error) {
	owner, repo, found := strings.Cut(id, "/")
	if !found {
		return "", errors.New("invalid repository id")
	}
	ctx := context.WithValue(context.Background(), github.SleepUntilPrimaryRateLimitResetWhenRateLimited, true)
	repository, _, err := c.client.Repositories.Get(ctx, owner, repo)
	isRateLimitError(err)
	if err != nil {
		return "", err
	}
	return c.LatestCommitHashForBranch(id, repository.GetDefaultBranch())
}

func (c *Client) LatestCommitHashForBranch(id, branch string) (string, error) {
	owner, repo, found := strings.Cut(id, "/")
	if !found {
		return "", errors.New("invalid repository id")
	}
	ctx := context.WithValue(context.Background(), github.SleepUntilPrimaryRateLimitResetWhenRateLimited, true)
	branchData, _, err := c.client.Repositories.GetBranch(ctx, owner, repo, branch, 10)
	isRateLimitError(err)
	if err != nil {
		return "", err
	}

	return branchData.GetCommit().GetSHA(), nil
}

func (c *Client) Checkout(path string, repository common.Repository) error {
	owner, repo, found := strings.Cut(repository.Id, "/")
	if !found {
		return errors.New("invalid repository id")
	}
	checkoutDirname := filepath.Join(path, c.RepositoryDirname(repository))
	if _, err := os.Stat(checkoutDirname); !os.IsNotExist(err) {
		// Already checked out, no need to download again
		return nil
	}

	ctx := context.WithValue(context.Background(), github.SleepUntilPrimaryRateLimitResetWhenRateLimited, true)

	// Gar tarball url
	url, _, err := c.client.Repositories.GetArchiveLink(ctx, owner, repo, github.Tarball, &github.RepositoryContentGetOptions{Ref: repository.CommitHash}, 10)
	isRateLimitError(err)
	if err != nil {
		return err
	}

	// Get tarball
	archiveResp, err := http.Get(url.String())
	if err != nil {
		return err
	}
	defer archiveResp.Body.Close()
	if archiveResp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code while downloading tarball: %d", archiveResp.StatusCode)
	}

	// Create temporary directory for extracting
	tmpDir, err := os.MkdirTemp(path, "tmp-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// Extract tarball
	tarCmd := exec.Command("tar", "-xzf", "-", "-C", tmpDir)
	tarPipe, err := tarCmd.StdinPipe()
	if err != nil {
		return err
	}
	err = tarCmd.Start()
	if err != nil {
		return err
	}
	_, err = io.Copy(tarPipe, archiveResp.Body)
	if err != nil {
		tarCmd.Process.Kill()
		return err
	}
	err = tarPipe.Close()
	if err != nil {
		tarCmd.Process.Kill()
		return err
	}
	err = tarCmd.Wait()
	if err != nil {
		return err
	}

	tmpDirEntries, err := os.ReadDir(tmpDir)
	if err != nil {
		return err
	}
	if len(tmpDirEntries) == 0 {
		return errors.New("no source code was extracted")
	}
	err = os.Rename(filepath.Join(tmpDir, tmpDirEntries[0].Name()), checkoutDirname)

	return nil
}

func (*Client) RepositoryDirname(repo common.Repository) string {
	owner, repoName, found := strings.Cut(repo.Id, "/")
	if !found {
		// Sorry
		panic("invalid repository id")
	}
	return owner + "#" + repoName + "#" + repo.CommitHash
}

func isRateLimitError(err error) bool {
	if errors.As(err, new(*github.RateLimitError)) {
		log.Println("hit rate limit")
		return true
	} else if errors.As(err, new(*github.AbuseRateLimitError)) {
		const sleepTime = 120
		log.Printf("hit secondary rate limit, sleeping for %d seconds\n", sleepTime)
		time.Sleep(sleepTime * time.Second)
		return true
	}
	return false
}
