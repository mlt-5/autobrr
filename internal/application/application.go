// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package application

import (
	"context"
	"encoding/json"
	"fmt"
	nethttp "net/http"

	"github.com/autobrr/autobrr/internal/action"
	"github.com/autobrr/autobrr/internal/alert"
	"github.com/autobrr/autobrr/internal/config"
	"github.com/autobrr/autobrr/internal/database"
	"github.com/autobrr/autobrr/internal/domain"
	"github.com/autobrr/autobrr/internal/events"
	"github.com/autobrr/autobrr/internal/feed"
	"github.com/autobrr/autobrr/internal/filter"
	"github.com/autobrr/autobrr/internal/http"
	"github.com/autobrr/autobrr/internal/indexer"
	"github.com/autobrr/autobrr/internal/irc"
	"github.com/autobrr/autobrr/internal/list"
	"github.com/autobrr/autobrr/internal/logger"
	"github.com/autobrr/autobrr/internal/meta"
	"github.com/autobrr/autobrr/internal/notification"
	"github.com/autobrr/autobrr/internal/proxy"
	"github.com/autobrr/autobrr/internal/release"
	"github.com/autobrr/autobrr/internal/scheduler"
	"github.com/autobrr/autobrr/pkg/errors"
	"github.com/autobrr/autobrr/pkg/featureflags"

	"github.com/KimMachineGun/automemlimit/memlimit"
	"github.com/dcarbone/zadapters/zstdlog"
	"github.com/r3labs/sse/v2"
	"github.com/rs/zerolog"
	"go.uber.org/automaxprocs/maxprocs"
)

func init() {
	featureflags.Register(domain.IRCFuzzyAnnouncer, false)
}

// App owns the lifecycle of every service, from config load to shutdown.
type App struct {
	log      zerolog.Logger
	cfg      *config.AppConfig
	sse      *sse.Server
	eventBus *events.EventBus

	db         *database.DB
	httpServer *http.Server

	handler nethttp.Handler
	ready   chan struct{}

	actionService       *action.Service
	alertService        *alert.Service
	indexerService      *indexer.Service
	indexerAPIService   *indexer.APIService
	ircService          *irc.Service
	feedService         *feed.Service
	filterService       *filter.Service
	notificationService *notification.Service
	releaseService      *release.Service
	rlsDownloadService  *release.DownloadService
	scheduler           *scheduler.Service
	listService         *list.Service
	proxyService        *proxy.Service
}

// New loads the config and builds the parts that need no database. The host,
// such as the iOS app, passes requests to Handler itself, so no network
// listener is opened.
func New(configPath string) *App {
	cfg := config.New(configPath)

	serverEvents := sse.New()
	serverEvents.CreateStreamWithOpts(logger.StreamLogs, sse.StreamOpts{MaxEntries: 1000, AutoReplay: true})
	serverEvents.CreateStreamWithOpts("irc", sse.StreamOpts{MaxEntries: 0, AutoReplay: false, AutoStream: true})
	serverEvents.CreateStreamWithOpts(notification.InboxStreamKey, sse.StreamOpts{MaxEntries: 0, AutoReplay: false})

	log := logger.New(cfg.Config, serverEvents)

	cfg.DynamicReload(log)

	return &App{
		log:      log,
		cfg:      cfg,
		sse:      serverEvents,
		eventBus: events.NewEventBus(log),
		ready:    make(chan struct{}),
	}
}

// Ready is closed once every component has started.
func (app *App) Ready() <-chan struct{} {
	return app.ready
}

// Handler serves the web UI and API. It is nil until Ready is closed.
func (app *App) Handler() nethttp.Handler {
	return app.handler
}

// Run starts the application and blocks until ctx is cancelled or a component
// fails to start, then shuts everything down. Cancelling ctx is a
// clean shutdown and returns nil, also while still starting. An App runs once,
// a host that restarts it builds a new one with New.
func (app *App) Run(ctx context.Context) error {
	// the event streams outlive a failed start, which never reaches the http
	// component that normally closes them
	defer app.sse.Close()

	if ctx.Err() != nil {
		return nil
	}

	if err := app.setup(ctx); err != nil {
		if ctx.Err() != nil {
			return nil
		}

		return err
	}

	defer func() {
		if err := app.db.Close(); err != nil {
			app.log.Error().Err(err).Msg("could not close database")
		}
	}()

	runCtx, fail := context.WithCancelCause(ctx)
	defer fail(nil)

	started, err := app.startComponents(runCtx, app.components())
	if err != nil {
		fail(err)
	} else {
		close(app.ready)
	}

	<-runCtx.Done()

	app.log.Info().Msg("shutting down")

	app.stopComponents(started)

	if ctx.Err() != nil {
		return nil
	}

	return context.Cause(runCtx)
}

func (app *App) components() []component {
	return []component{
		{
			name:  "scheduler",
			start: app.scheduler.Start,
			stop:  app.scheduler.Stop,
		},
		{
			name:     "notifications",
			optional: true,
			start: func(context.Context) error {
				return app.notificationService.Start()
			},
		},
		{
			name: "indexers",
			start: func(context.Context) error {
				return app.indexerService.Start()
			},
		},
		{
			name: "irc",
			start: func(context.Context) error {
				app.ircService.StartHandlers()

				return nil
			},
			stop: func(context.Context) error {
				app.ircService.StopHandlers()

				return nil
			},
		},
		{
			name:     "feeds",
			optional: true,
			start: func(context.Context) error {
				return app.feedService.Start()
			},
		},
		{
			name:     "release cleanup",
			optional: true,
			start: func(context.Context) error {
				return app.releaseService.StartCleanupJobs()
			},
		},
		{
			name: "lists",
			start: func(context.Context) error {
				go func() {
					if err := app.listService.Start(); err != nil {
						app.log.Error().Err(err).Msg("could not start list service")
					}
				}()

				return nil
			},
		},
		// last to start so it stops first, before the services its handlers call
		{
			name: "http",
			start: func(context.Context) error {
				app.handler = app.httpServer.Handler()

				return nil
			},
			// ends open event streams, which would otherwise outlive the services
			stop: func(context.Context) error {
				app.sse.Close()

				return nil
			},
		},
	}
}

// Config returns the loaded configuration.
func (app *App) Config() *domain.Config {
	return app.cfg.Config
}

// Log returns the application logger.
func (app *App) Log() *zerolog.Logger {
	return &app.log
}

func (app *App) setupResourceLimits() {
	// the returned undo is dropped on purpose, the limit holds for the process lifetime
	if _, err := maxprocs.Set(maxprocs.Logger(zstdlog.NewStdLoggerWithLevel(app.log.With().Logger(), zerolog.InfoLevel).Printf)); err != nil {
		app.log.Error().Err(err).Msg("failed to set GOMAXPROCS")
	}

	memLimit, err := memlimit.Set(memlimit.WithProvider(memlimit.ApplyFallback(memlimit.FromCgroup, memlimit.FromSystem)))
	if err != nil {
		app.log.Error().Err(err).Msg("failed to set GOMEMLIMIT")
	}

	app.log.Debug().Int64("gomemlimit_bytes", memLimit).Msg("memory limit configured")
}

func (app *App) setup(ctx context.Context) error {
	app.setupResourceLimits()

	db, err := database.NewDB(app.cfg.Config, app.log)
	if err != nil {
		return errors.Wrap(err, "could not init database")
	}

	if err := db.Open(); err != nil {
		return errors.Wrap(err, "could not open database")
	}

	app.db = db

	app.log.Info().
		Str("version", meta.GetVersion()).
		Str("commit", meta.GetCommit()).
		Str("build_date", meta.GetDate()).
		Str("log_level", app.cfg.Config.LogLevel).
		Msg("starting autobrr")

	var (
		actionRepo       = database.NewActionRepo(app.log, db)
		filterRepo       = database.NewFilterRepo(app.log, db)
		feedRepo         = database.NewFeedRepo(app.log, db)
		feedCacheRepo    = database.NewFeedCacheRepo(app.log, db)
		indexerRepo      = database.NewIndexerRepo(app.log, db)
		ircRepo          = database.NewIrcRepo(app.log, db)
		listRepo         = database.NewListRepo(app.log, db)
		notificationRepo = database.NewNotificationRepo(app.log, db)
		inboxRepo        = database.NewNotificationInboxRepo(app.log, db)
		releaseRepo      = database.NewReleaseRepo(app.log, db)
		proxyRepo        = database.NewProxyRepo(app.log, db)
	)

	app.scheduler = scheduler.NewService(app.log)
	app.notificationService = notification.NewService(app.log, app.eventBus, app.sse, notificationRepo, inboxRepo, app.scheduler)
	app.proxyService = proxy.NewService(app.log, app.eventBus, proxyRepo)
	app.indexerAPIService = indexer.NewAPIService(app.log, app.proxyService)
	app.rlsDownloadService = release.NewDownloadService(app.log, indexerRepo, app.proxyService)
	app.actionService = action.NewService(app.log, app.eventBus, actionRepo, app.rlsDownloadService)
	app.indexerService = indexer.NewService(app.log, app.eventBus, app.cfg.Config, indexerRepo, releaseRepo, app.indexerAPIService)
	app.filterService = filter.NewService(app.log, filterRepo, app.actionService, releaseRepo, app.indexerAPIService, app.indexerService, app.rlsDownloadService, app.notificationService)
	app.releaseService = release.NewService(app.log, app.eventBus, releaseRepo, app.actionService, app.filterService, app.indexerService, app.scheduler)
	app.ircService = irc.NewService(app.log, app.eventBus, app.sse, ircRepo, app.releaseService, app.indexerService, app.proxyService)
	app.feedService = feed.NewService(app.log, app.eventBus, feedRepo, feedCacheRepo, app.releaseService, app.proxyService, app.scheduler)
	app.listService = list.NewService(app.log, app.eventBus, listRepo, app.filterService, app.scheduler)
	app.alertService = alert.NewService(app.log, app.eventBus, app.sse, app.ircService, app.listService)

	app.httpServer = http.NewServer(http.Deps{
		Log:                 app.log,
		SSE:                 app.sse,
		DB:                  db,
		Config:              app.cfg,
		Version:             meta.GetVersion(),
		Commit:              meta.GetCommit(),
		Date:                meta.GetDate(),
		ActionService:       app.actionService,
		AlertService:        app.alertService,
		FilterService:       app.filterService,
		FeedService:         app.feedService,
		IndexerService:      app.indexerService,
		IrcService:          app.ircService,
		ListService:         app.listService,
		NotificationService: app.notificationService,
		ProxyService:        app.proxyService,
		ReleaseService:      app.releaseService,
	})

	return nil
}

func PrintVersion() {
	fmt.Println(meta.GetMetaStr())
}

func PrintVersionJSON() {
	bytes, err := json.Marshal(meta.GetMetaInfo())
	if err != nil {
		return
	}

	fmt.Println(string(bytes))
}
