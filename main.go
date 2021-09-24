package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/v39/github"
)

// GithubRepoService is an interface to help with dependency injection for extensible,
// testable code. Matches methods from go-github's type RepositoriesService
type GithubRepoService interface {
	CompareCommits(ctx context.Context, owner, repo string, base, head string, options *github.ListOptions) (*github.CommitsComparison, *github.Response, error)
}

// GithubClient wraps the github API client to help with mocking responses in tests
type GithubClient struct {
	service GithubRepoService
	config  ClientConfig
}

// ClientConfig specifies required parameters for commit comparison requests against GitHub
type ClientConfig struct {
	owner string
	repo  string
	head  string
	base  string
	user  string
	pass  string
}

// Control specifies optional output control parameters for end users
type Control struct {
	json bool
}

// CommitMessage specifies the message content, author, and associated git commit
// sha of commits to be fetched in the comparison.
type CommitMessage struct {
	Author    *github.CommitAuthor `json:"author,omitempty"`
	CommitSHA *string              `json:"sha,omitempty"`
	Message   *string              `json:"message,omitempty"`
}

// NewGithubClient constructs a GithubClient, specifying the client only. The remainder of
// settings are set by command-line flags and applied/parsed in main
func NewGithubClient(svc GithubRepoService, cfg ClientConfig) *GithubClient {
	return &GithubClient{
		service: svc,
		config:  cfg,
	}
}

// GetCommitMessages calls the GitHub API for comparison of commits between specified BASE and HEAD
// It returns commit messages, SHAs, and author info as a byte slice
func GetCommitMessages(client *GithubClient) ([]CommitMessage, error) {

	ctx := context.Background()

	var res []CommitMessage

	for page, numCommits, firstRun := 1, 1, true; numCommits > 0; page++ {
		options := &github.ListOptions{Page: page, PerPage: 100}
		c, _, err := client.service.CompareCommits(ctx, client.config.owner, client.config.repo, client.config.base, client.config.head, options)
		if err != nil {
			return nil, err
		}

		for i := range c.Commits {
			currentCommit := c.Commits[i].Commit

			m := CommitMessage{
				Author:    currentCommit.Author,
				CommitSHA: currentCommit.Tree.SHA,
				Message:   currentCommit.Message,
			}

			res = append(res, m)
		}

		if firstRun == true {
			numCommits = *c.TotalCommits

			firstRun = false
		}

		numCommits = numCommits - 100
	}

	return res, nil
}

func printCommitMessages(ctl *Control, c []CommitMessage) error {

	var res []string

	if ctl.json == true {

		j, err := json.MarshalIndent(c, "", " ")
		if err != nil {
			return err
		}

		fmt.Println(string(j))
		return nil
	}

	for i := range c {
		formatted := fmt.Sprintf(
			"Date: %s\n"+
				"Commit: %s \n"+
				"Author: %s (%s)\n"+
				"Message: %s\n\n",
			*c[i].Author.Date,
			*c[i].CommitSHA,
			*c[i].Author.Name,
			*c[i].Author.Email,
			*c[i].Message)

		res = append(res, formatted)
	}

	out := strings.Trim(fmt.Sprint(res), "[]")
	fmt.Println(out)

	return nil
}

func main() {
	cfg := new(ClientConfig)
	ctl := new(Control)

	flag.StringVar(&cfg.user, "user", "", "User to authenticate as in GitHub")
	flag.StringVar(&cfg.pass, "pass", "", "Password or Personal Access Token to use for authentication in GitHub")
	flag.StringVar(&cfg.owner, "owner", "", "Name of the user or organization the repo belongs to")
	flag.StringVar(&cfg.repo, "repo", "", "Name of the repo to compare commits from")
	flag.StringVar(&cfg.head, "head", "", "Git commit SHA, branch, or tag of commit to use as HEAD in the comparison")
	flag.StringVar(&cfg.base, "base", "", "Git commit SHA, branch, or tag of commit to use as BASE in the comparison")
	flag.BoolVar(&ctl.json, "json", false, "Output content in JSON format")
	flag.Parse()

	tp := github.BasicAuthTransport{
		Username: cfg.user,
		Password: cfg.pass,
	}

	gh := github.NewClient(tp.Client())

	client := NewGithubClient(gh.Repositories, *cfg)

	res, err := GetCommitMessages(client)
	if err != nil {
		log.Panicln(err)
	}

	err = printCommitMessages(ctl, res)
	if err != nil {
		log.Panicln(err)
	}
}
