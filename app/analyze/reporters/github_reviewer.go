package reporters

import (
	"context"
	"fmt"

	"github.com/golangci/golangci-worker/app/analyze/linters/result"
	"github.com/golangci/golangci-worker/app/utils/github"
	gh "github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
)

type GithubReviewer struct {
	*github.Context
	client github.Client
}

func NewGithubReviewer(c *github.Context, client github.Client) *GithubReviewer {
	return &GithubReviewer{
		Context: c,
		client:  client,
	}
}

func (gr GithubReviewer) Report(ctx context.Context, ref string, issues []result.Issue) error {
	if len(issues) == 0 {
		logrus.Infof("Nothing to report")
		return nil
	}

	comments := []*gh.DraftReviewComment{}
	for _, i := range issues {
		comment := &gh.DraftReviewComment{
			Path:     gh.String(i.File),
			Position: gh.Int(i.HunkPos),
			Body:     gh.String(i.Text),
		}
		comments = append(comments, comment)
	}

	review := &gh.PullRequestReviewRequest{
		CommitID: gh.String(ref),
		Event:    gh.String("COMMENT"),
		Body:     gh.String(""),
		Comments: comments,
	}
	if err := gr.client.CreateReview(ctx, gr.Context, review); err != nil {
		return fmt.Errorf("can't create review %+v: %s", review, err)
	}

	logrus.Infof("Submitted review %+v", review)
	return nil
}
