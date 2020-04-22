package bencher

import (
	"context"
	"time"

	"github.com/victor-carvalho/gobench/config"
	"github.com/victor-carvalho/gobench/requester"
	"github.com/victor-carvalho/gobench/stats"
)

type Bencher struct {
	cfg       config.Config
	requester *requester.Requester
}

func NewFromConfig(cfg config.Config) *Bencher {
	return &Bencher{
		cfg:       cfg,
		requester: requester.NewRequester(cfg),
	}
}

func (b *Bencher) Run() stats.Stats {
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(b.cfg.MaxElapsedTime))
	go b.requester.Run(ctx)
	return stats.CollectStats(ctx, b.cfg, b.requester.Output())
}
