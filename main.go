package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/google/go-github/v39/github"
)

// Request specifies the parameters required for commit comparison requests against
// GitHub, as well as any additional formatting/output control settings
type Request struct {
	owner string
	repo  string
	head  string
	base  string
	user  string
	pass  string
	short bool
}

// CommitMessage specifies the message content, author, and associated git commit
// sha of commits to be fetched in the comparison.
type CommitMessage struct {
	Author    *github.CommitAuthor `json:"author,omitempty"`
	CommitSHA *string              `json:"sha,omitempty"`
	Message   *string              `json:"message,omitempty"`
}

// GetCommitMessages calls the GitHub API for comparison of commits between specified BASE and HEAD
// It returns commit messages, SHAs, and author info as a byte slice
func GetCommitMessages(r *Request) (string, error) {
	//func GetCommitMessages(c *Comparison) ([]byte, error) {

	ctx := context.Background()

	tp := github.BasicAuthTransport{
		Username: r.user,
		Password: r.pass,
	}

	client := github.NewClient(tp.Client())

	var res []CommitMessage

	for page, numCommits, firstRun := 1, 1, true; numCommits > 0; page++ {
		options := &github.ListOptions{page, 100}
		c, _, err := client.Repositories.CompareCommits(ctx, r.owner, r.repo, r.base, r.head, options)
		if err != nil {
			return "Error", err
		}

		for i, l := 0, len(c.Commits); i < l-1; i++ {
			currentCommit := c.Commits[i].Commit

			m := CommitMessage{
				Author:    currentCommit.Author,
				CommitSHA: currentCommit.SHA,
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

	if r.short {
		fmt.Println("Short input was requested")
	}

	resJSON, err := json.MarshalIndent(res, "", " ")
	if err != nil {
		return "Error", err
	}

	return string(resJSON), nil
}

func main() {
	c := new(Request)
	flag.StringVar(&c.owner, "owner", "", "Name of the user or organization the repo belongs to")
	flag.StringVar(&c.repo, "repo", "", "Name of the repo to compare commits from")
	flag.StringVar(&c.head, "head", "", "Git commit SHA, branch, or tag of commit to use as HEAD in the comparison")
	flag.StringVar(&c.base, "base", "", "Git commit SHA, branch, or tag of commit to use as BASE in the comparison")
	flag.StringVar(&c.user, "user", "", "User to authenticate as in GitHub")
	flag.StringVar(&c.pass, "pass", "", "Password or Personal Access Token to use for authentication in GitHub")
	flag.BoolVar(&c.short, "short", false, "Minimal output with messages only (omit author and commit info)")
	flag.Parse()

	res, err := GetCommitMessages(c)
	if err != nil {
		log.Panicln(err)
	}

	fmt.Println(res)
}
