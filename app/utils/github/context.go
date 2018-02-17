package github

import (
	"context"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Repo struct {
	Owner, Name string
}

type Context struct {
	Repo              Repo
	GithubAccessToken string
	PullRequestNumber int
}

func (c Context) GetClient(ctx context.Context) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.GithubAccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

var FakeContext = Context{
	Repo: Repo{
		Owner: "owner",
		Name:  "name",
	},
	GithubAccessToken: "access_token",
	PullRequestNumber: 1,
}
