package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/google/go-github/v39/github"
)

// Request specifies required parameters for commit comparison requests against GitHub
type Request struct {
	owner string
	repo  string
	head  string
	base  string
	user  string
	pass  string
}

// control specifies optional output control parameters for end users
type control struct {
	json bool
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
func GetCommitMessages(req *Request) ([]CommitMessage, []byte, error) {

	ctx := context.Background()

	tp := github.BasicAuthTransport{
		Username: req.user,
		Password: req.pass,
	}

	client := github.NewClient(tp.Client())

	var res []CommitMessage

	for page, numCommits, firstRun := 1, 1, true; numCommits > 0; page++ {
		options := &github.ListOptions{Page: page, PerPage: 100}
		c, _, err := client.Repositories.CompareCommits(ctx, req.owner, req.repo, req.base, req.head, options)
		if err != nil {
			return nil, nil, err
		}

		for i, l := 0, len(c.Commits); i < l-1; i++ {
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

	resJSON, err := json.MarshalIndent(res, "", " ")
	if err != nil {
		return nil, nil, err
	}

	return res, resJSON, nil
}

func printCommitMessages(ctl *control, c []CommitMessage, m []byte) error {

	var res []string

	if ctl.json == true {
		fmt.Println(string(m))
		return nil
	}

	for i := range c {
		formatted := fmt.Sprintf(
			"Date: %s\n"+
				"  Commit: %s \n"+
				"  Author: %s (%s)\n"+
				"  Message: %s\n\n",
			*c[i].Author.Date,
			*c[i].CommitSHA,
			*c[i].Author.Name,
			*c[i].Author.Email,
			*c[i].Message)

		res = append(res, formatted)
	}

	fmt.Println(res)

	return nil
}

func main() {
	r := new(Request)
	c := new(control)
	flag.StringVar(&r.owner, "owner", "", "Name of the user or organization the repo belongs to")
	flag.StringVar(&r.repo, "repo", "", "Name of the repo to compare commits from")
	flag.StringVar(&r.head, "head", "", "Git commit SHA, branch, or tag of commit to use as HEAD in the comparison")
	flag.StringVar(&r.base, "base", "", "Git commit SHA, branch, or tag of commit to use as BASE in the comparison")
	flag.StringVar(&r.user, "user", "", "User to authenticate as in GitHub")
	flag.StringVar(&r.pass, "pass", "", "Password or Personal Access Token to use for authentication in GitHub")
	flag.BoolVar(&c.json, "json", false, "Output content in JSON format")
	flag.Parse()

	res, resJSON, err := GetCommitMessages(r)
	if err != nil {
		log.Panicln(err)
	}

	err = printCommitMessages(c, res, resJSON)
	if err != nil {
		log.Panicln(err)
	}
}
