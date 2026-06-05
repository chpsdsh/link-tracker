package metrics

import (
	"context"
	"log/slog"
	"time"
)

const (
	updateRequestDuration = 15 * time.Second
	githubLabel           = "github"
	stackOverflowLabel    = "stackoverflow"
)

type LinksCountRequester interface {
	CountLinksOnTrack(ctx context.Context) (int, int, error)
}

type LinksOnTrackUpdater struct {
	Requester LinksCountRequester
	Logger    *slog.Logger
}

func (u LinksOnTrackUpdater) UpdateLinksCount() {
	ctx, cancel := context.WithTimeout(context.Background(), updateRequestDuration)
	defer cancel()

	gitLinks, stackLinks, err := u.Requester.CountLinksOnTrack(ctx)
	if err != nil {
		u.Logger.Error("error counting links")
		return
	}
	LinksOnTrackTotal.
		WithLabelValues(githubLabel).
		Set(float64(gitLinks))
	LinksOnTrackTotal.
		WithLabelValues(stackOverflowLabel).
		Set(float64(stackLinks))
}
