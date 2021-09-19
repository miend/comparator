package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/google/go-github/v39/github"
)

// Comparison specifies the parameters required for commit comparison requests against
// GitHub, as well as any additional formatting/output control settings
type Comparison struct {
	owner string
	repo  string
	head  string
	base  string
	user  string
	pass  string
	short bool
}

// GetCommitMessages calls the GitHub API for comparison of commits between specified BASE and HEAD
// It returns commit messages, SHAs, and author info as a byte slice
func GetCommitMessages(c *Comparison) (string, error) {
	//func GetCommitMessages(c *Comparison) ([]byte, error) {

	ctx := context.Background()

	tp := github.BasicAuthTransport{
		Username: c.user,
		Password: c.pass,
	}

	client = github.NewClient(tp.Client())

	listOptions := &github.ListOptions{1, 100}

	r, _, err := client.Repositories.CompareCommits(ctx, c.owner, c.repo, c.base, c.head, listOptions)
	if err != nil {
		return "Error", err
	}

	if c.short {
		fmt.Println("Short input was requested")
	}

	fmt.Println(r)

	return "Test", nil
}

func main() {
	c := new(Comparison)
	flag.StringVar(&c.owner, "owner", "", "Name of the user or organization the repo belongs to")
	flag.StringVar(&c.repo, "repo", "", "Name of the repo to compare commits from")
	flag.StringVar(&c.head, "head", "", "Git commit SHA, branch, or tag of commit to use as HEAD in the comparison")
	flag.StringVar(&c.base, "base", "", "Git commit SHA, branch, or tag of commit to use as BASE in the comparison")
	flag.StringVar(&c.user, "user", "", "User to authenticate as in GitHub")
	flag.StringVar(&c.pass, "pass", "", "Password or Personal Access Token to use for authentication in GitHub")
	flag.BoolVar(&c.short, "short", false, "Minimal output with messages only (omit author and commit info)")
	flag.Parse()

	GetCommitMessages(c)
}
