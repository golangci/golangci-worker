package httputils

import (
	"context"
	"fmt"
	"io"

	"github.com/golangci/golangci-worker/app/analytics"
	"github.com/levigross/grequests"
)

type Client interface {
	Get(ctx context.Context, url string) (io.ReadCloser, error)
}

type GrequestsClient struct{}

func (c GrequestsClient) Get(ctx context.Context, url string) (io.ReadCloser, error) {
	resp, err := grequests.Get(url, &grequests.RequestOptions{
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to make http request %q: %s", url, err)
	}

	if !resp.Ok {
		if cerr := resp.Close(); cerr != nil {
			analytics.Log(ctx).Warnf("Can't close %q response: %s", url, cerr)
		}

		return nil, fmt.Errorf("got error code from %q: %d", url, resp.StatusCode)
	}

	return resp, nil
}
