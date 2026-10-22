// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package http

import (
	"net/http"

	"github.com/autobrr/autobrr/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/r3labs/sse/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

type Server struct {
	log zerolog.Logger
	sse *sse.Server
	db  DatabaseHealth

	buildInfo buildInfo
	config    *config.AppConfig

	actionService       actionService
	alertService        alertService
	filterService       filterService
	feedService         feedService
	indexerService      indexerService
	ircService          ircService
	listService         listService
	notificationService notificationService
	proxyService        proxyService
	releaseService      releaseService
}

type Deps struct {
	Log zerolog.Logger
	SSE *sse.Server
	DB  DatabaseHealth

	Config *config.AppConfig

	Version string
	Commit  string
	Date    string

	ActionService       actionService
	AlertService        alertService
	FilterService       filterService
	FeedService         feedService
	IndexerService      indexerService
	IrcService          ircService
	ListService         listService
	NotificationService notificationService
	ProxyService        proxyService
	ReleaseService      releaseService
}

func NewServer(deps Deps) *Server {
	srv := &Server{
		log:    deps.Log.With().Str("module", "http").Logger(),
		config: deps.Config,
		sse:    deps.SSE,
		db:     deps.DB,
		buildInfo: buildInfo{
			version: deps.Version,
			commit:  deps.Commit,
			date:    deps.Date,
		},

		actionService:       deps.ActionService,
		alertService:        deps.AlertService,
		filterService:       deps.FilterService,
		feedService:         deps.FeedService,
		indexerService:      deps.IndexerService,
		ircService:          deps.IrcService,
		listService:         deps.ListService,
		notificationService: deps.NotificationService,
		proxyService:        deps.ProxyService,
		releaseService:      deps.ReleaseService,
	}

	return srv
}

// Handler serves the API under /api.
func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()

	r.Use(hlog.NewHandler(s.log))
	r.Use(hlog.RequestIDHandler("request_id", "X-Request-Id"))
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(LoggerMiddleware(&s.log))

	encoder := newEncoder(s.log)

	// Create a separate router for API
	apiRouter := chi.NewRouter()
	apiRouter.Route("/healthz", newHealthHandler(encoder, s.db).Routes)
	apiRouter.Route("/actions", newActionHandler(encoder, s.actionService).Routes)
	apiRouter.Route("/alerts", newAlertHandler(encoder, s.alertService).Routes)
	apiRouter.Route("/config", newConfigHandler(encoder, s.buildInfo, s.config).Routes)
	apiRouter.Route("/filters", newFilterHandler(encoder, s.filterService).Routes)
	apiRouter.Route("/feeds", newFeedHandler(encoder, s.feedService).Routes)
	apiRouter.Route("/irc", newIrcHandler(encoder, s.sse, s.ircService).Routes)
	apiRouter.Route("/indexer", newIndexerHandler(encoder, s.indexerService, s.ircService).Routes)
	apiRouter.Route("/lists", newListHandler(encoder, s.listService).Routes)
	apiRouter.Route("/logs", newLogsHandler(s.config).Routes)
	apiRouter.Route("/notification", newNotificationHandler(encoder, s.notificationService).Routes)
	apiRouter.Route("/proxy", newProxyHandler(encoder, s.proxyService).Routes)
	apiRouter.Route("/release", newReleaseHandler(encoder, s.releaseService).Routes)

	apiRouter.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		s.sse.Headers = map[string]string{
			"Content-Type":      "text/event-stream",
			"Cache-Control":     "no-cache",
			"Connection":        "keep-alive",
			"X-Accel-Buffering": "no",
		}

		s.sse.ServeHTTP(w, r)
	})

	r.Mount("/api", apiRouter)

	return r
}
