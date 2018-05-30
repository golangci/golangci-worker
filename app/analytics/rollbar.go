package analytics

import (
	"context"
	"os"

	"github.com/golangci/golangci-worker/app/utils/runmode"
	"github.com/stvp/rollbar"
)

func trackError(ctx context.Context, err error, level string) {
	if !runmode.IsProduction() {
		return
	}

	trackingProps := getTrackingProps(ctx)
	tf := &rollbar.Field{
		Name: "props",
		Data: trackingProps,
	}
	pf := &rollbar.Field{
		Name: "project",
		Data: "worker",
	}

	rollbar.Error(level, err, tf, pf)
}

func init() {
	rollbar.Token = os.Getenv("ROLLBAR_API_TOKEN")
	rollbar.Environment = "production" // defaults to "development"
}
