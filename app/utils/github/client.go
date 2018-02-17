package github

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/github"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusFailure Status = "failure"
	StatusError   Status = "error"
	StatusSuccess Status = "success"
)

type Client interface {
	GetPullRequest(ctx context.Context, c *Context) (*gh.PullRequest, error)
	GetPullRequestPatch(ctx context.Context, c *Context) (string, error)
	CreateReview(ctx context.Context, c *Context, review *gh.PullRequestReviewRequest) error
	SetCommitStatus(ctx context.Context, c *Context, ref string, status Status, desc string) error
}

type MyClient struct{}

var _ Client = &MyClient{}

func NewMyClient() *MyClient {
	return &MyClient{}
}

func (gc *MyClient) GetPullRequest(ctx context.Context, c *Context) (*gh.PullRequest, error) {
	ghClient := c.GetClient(ctx)
	pr, _, err := ghClient.PullRequests.Get(ctx, c.Repo.Owner, c.Repo.Name, c.PullRequestNumber)
	if err != nil {
		return nil, fmt.Errorf("can't get pull request %d from github: %s", c.PullRequestNumber, err)
	}

	return pr, nil
}

func (gc *MyClient) CreateReview(ctx context.Context, c *Context, review *gh.PullRequestReviewRequest) error {
	_, _, err := c.GetClient(ctx).PullRequests.CreateReview(ctx, c.Repo.Owner, c.Repo.Name, c.PullRequestNumber, review)
	if err != nil {
		return fmt.Errorf("can't create github review: %s", err)
	}

	return nil
}

func (gc *MyClient) GetPullRequestPatch(ctx context.Context, c *Context) (string, error) {
	opts := gh.RawOptions{Type: gh.Patch}
	raw, _, err := c.GetClient(ctx).PullRequests.GetRaw(ctx, c.Repo.Owner, c.Repo.Name,
		c.PullRequestNumber, opts)
	if err != nil {
		return "", fmt.Errorf("can't get patch for pull request: %s", err)
	}

	return raw, nil
}

func (gc *MyClient) SetCommitStatus(ctx context.Context, c *Context, ref string, status Status, desc string) error {
	rs := &gh.RepoStatus{
		Description: gh.String(desc),
		State:       gh.String(string(status)),
		Context:     gh.String("GolangCI"),
	}
	_, _, err := c.GetClient(ctx).Repositories.CreateStatus(ctx, c.Repo.Owner, c.Repo.Name, ref, rs)
	if err != nil {
		return fmt.Errorf("can't set commit %s status %s: %s", ref, status, err)
	}

	return nil
}
