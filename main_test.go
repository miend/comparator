package main

import (
	"context"
	"github.com/google/go-github/v39/github"
	"reflect"
	"testing"
	"time"
)

// Set up things required for mocked responses instead of real GitHub API calls
var testTime = time.Date(2021, 9, 23, 1, 10, 58, 260198254, time.UTC)
var testUser = "Test User"
var testEmail = "test@example.com"
var testSHA = "abc123"
var testMessage = "I'm a test commit / I exist for one purpose / Something about bugs"

var testAuthor = github.CommitAuthor{
	Date:  &testTime,
	Name:  &testUser,
	Email: &testEmail,
	// we don't use Login for anything
}

var testCommitMessage = &CommitMessage{
	Author:    &testAuthor,
	CommitSHA: &testSHA,
	Message:   &testMessage,
}

var testConfig = &ClientConfig{
	owner: "test",
	repo:  "test",
	head:  "test",
	base:  "test",
	user:  "test",
	pass:  "test",
}

// Determines how many commits will be tested in a comparison
// Defined globally because I don't see a good way to pass this
// to CompareCommits otherwise.
// Set it just before calling CompareCommits
// TODO: Find a better way to do this
var testCommitCount int

type MockGithubRepoService struct{}

func (f MockGithubRepoService) CompareCommits(ctx context.Context, owner, repo string, base, head string, options *github.ListOptions) (*github.CommitsComparison, *github.Response, error) {
	comp := generateComparison(testCommitCount)
	// TODO: investigate options for testing various responses
	res := &github.Response{}
	return comp, res, nil
}

// Test as many commits in a comparison as you think you need
func generateComparison(n int) *github.CommitsComparison {

	t := github.Tree{
		SHA: &testSHA,
	}

	c := github.Commit{
		Author:  &testAuthor,
		Message: &testMessage,
		Tree:    &t,
	}

	r := github.RepositoryCommit{
		Commit: &c,
	}

	commits := make([]*github.RepositoryCommit, n)

	for i := range commits {
		commits[i] = &r
	}

	res := &github.CommitsComparison{
		TotalCommits: &n,
		Commits:      commits,
	}

	return res
}

// Begin actual tests
func TestGetCommitMessages(t *testing.T) {

	client := NewGithubClient(new(MockGithubRepoService), *testConfig)

	// Using any int above 100 breaks testing because the mock scenario can't emulate
	// the real Github API's pagination, and CompareCommits here will always return the
	// same number of commits... starting to think just using the real Github API would've
	// been a better idea than trying to mock it?
	sets := []int{3, 50}

	// Mocking responses means we don't actually test that pagination is functioning...
	for c := range sets {

		testCommitCount = sets[c]

		want := make([]CommitMessage, testCommitCount)

		for i := range want {
			want[i] = *testCommitMessage
		}

		got, err := GetCommitMessages(client)
		if err != nil {
			t.Fatalf("\nExpected: %v\n\nGot: %v", want, got)
		}

		if !reflect.DeepEqual(want, got) {
			t.Fatalf("\nExpected: %v\n\nGot: %v", want, got)
		}
	}
}

func TestPrintCommitMessages(t *testing.T) {

	messages := make([]CommitMessage, 3)

	for i := range messages {
		messages[i] = *testCommitMessage
	}

	// want nil because the function only returns errors (if they occur)
	type test struct {
		json bool
		want error
	}

	// It looks unnecessary, but using a table-driven test makes this easier to
	// extend later.
	tests := []test{
		{json: true, want: nil},
		{json: false, want: nil},
	}

	for _, testCase := range tests {

		ctl := &Control{
			json: testCase.json,
		}

		got := printCommitMessages(ctl, messages)

		if testCase.want != got {
			t.Fatalf("Expected %v, got %v", testCase.want, got)
		}
	}
}
