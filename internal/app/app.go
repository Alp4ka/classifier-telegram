package app

import (
	"context"
	"sync"
	"time"

	"github.com/Alp4ka/classifier-telegram/internal/interactions/core"
	"github.com/go-telegram/bot"
	"github.com/google/uuid"
	"github.com/jellydator/ttlcache/v3"
)

type App struct {
	cfg    Config
	bot    *bot.Bot
	client core.Client
	cache  *ttlcache.Cache[string, uuid.UUID]
	mu     sync.Mutex
}

func New(cfg Config) (*App, error) {
	a := App{
		client: cfg.CoreClient,
		cfg:    cfg,
	}

	b, err := bot.New(cfg.APIKey, bot.WithDefaultHandler(a.handler))
	if err != nil {
		return nil, err
	}
	a.bot = b

	return &a, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	a.cache = ttlcache.New[string, uuid.UUID](
		ttlcache.WithTTL[string, uuid.UUID](5 * time.Minute),
	)
	go a.cache.Start()
	go a.bot.Start(ctx)

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

func (a *App) Close() (err error) {
	return nil
}
