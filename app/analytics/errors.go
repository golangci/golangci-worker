package analytics

import (
	"context"

	"github.com/golangci/golangci-shared/pkg/config"
	"github.com/golangci/golangci-shared/pkg/logutil"

	"github.com/golangci/golangci-shared/pkg/apperrors"
	"github.com/golangci/golangci-worker/app/lib/runmode"
)

func trackError(ctx context.Context, err error, level apperrors.Level) {
	if !runmode.IsProduction() {
		return
	}

	log := logutil.NewStderrLog("trackError")
	cfg := config.NewEnvConfig(log)
	et := apperrors.GetTracker(cfg, log, "worker")
	et.Track(level, err.Error(), getTrackingProps(ctx))
}
