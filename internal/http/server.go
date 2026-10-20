// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/autobrr/autobrr/internal/config"
	"github.com/autobrr/autobrr/pkg/errors"
	"github.com/autobrr/autobrr/web"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/r3labs/sse/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

type Server struct {
	log      zerolog.Logger
	sse      *sse.Server
	db       DatabaseHealth
	server   *http.Server
	listener net.Listener

	buildInfo      buildInfo
	config         *config.AppConfig
	sessionManager *scs.SessionManager

	actionService       actionService
	alertService        alertService
	authService         authService
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

	Config         *config.AppConfig
	SessionManager *scs.SessionManager

	Version string
	Commit  string
	Date    string

	ActionService       actionService
	AlertService        alertService
	AuthService         authService
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
	sessionManager := deps.SessionManager
	sessionManager.Lifetime = 24 * time.Hour * 30
	sessionManager.Cookie.Name = "autobrr_user_session"
	sessionManager.Cookie.Persist = false
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Path = "/"
	sessionManager.Cookie.Secure = false

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

		sessionManager: sessionManager,

		actionService:       deps.ActionService,
		alertService:        deps.AlertService,
		authService:         deps.AuthService,
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

// Listen binds the configured address. It returns before serving so a port
// conflict fails startup instead of surfacing later.
func (s *Server) Listen() error {
	listener, err := listen(s.log, fmt.Sprintf("%v:%v", s.config.Config.Host, s.config.Config.Port))
	if err != nil {
		return err
	}

	s.listener = listener
	s.server = &http.Server{
		Handler:           s.Handler(),
		ReadHeaderTimeout: time.Second * 15,
	}

	s.log.Info().Str("addr", listener.Addr().String()).Msg("starting api server")

	return nil
}

// Serve blocks until the server is shut down. It requires a prior Listen.
func (s *Server) Serve() error {
	if err := s.server.Serve(s.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

// Shutdown ends the event streams first, since SSE handlers only return when
// their stream closes and would otherwise hold Shutdown until ctx expires.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	s.sse.Close()

	if err := s.server.Shutdown(ctx); err != nil {
		_ = s.server.Close()

		return errors.Wrap(err, "could not shut down gracefully")
	}

	return nil
}

// CloseListener releases the port when Serve never ran. Once Serve has run,
// Shutdown already closed the listener and this is a no-op.
func (s *Server) CloseListener(context.Context) error {
	if s.listener == nil {
		return nil
	}

	if err := s.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return err
	}

	return nil
}

func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()

	r.Use(hlog.NewHandler(s.log))
	r.Use(hlog.RequestIDHandler("request_id", "X-Request-Id"))
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(LoggerMiddleware(&s.log))

	r.Use(s.sessionMiddleware)

	encoder := newEncoder(s.log)

	// Create a separate router for API
	apiRouter := chi.NewRouter()
	apiRouter.Route("/auth", newAuthHandler(encoder, s.log, s, s.config.Config, s.sessionManager, s.authService).Routes)
	apiRouter.Route("/healthz", newHealthHandler(encoder, s.db).Routes)
	apiRouter.Group(func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(s.IsAuthenticated)

			r.Route("/actions", newActionHandler(encoder, s.actionService).Routes)
			r.Route("/alerts", newAlertHandler(encoder, s.alertService).Routes)
			r.Route("/config", newConfigHandler(encoder, s.buildInfo, s.config).Routes)
			r.Route("/filters", newFilterHandler(encoder, s.filterService).Routes)
			r.Route("/feeds", newFeedHandler(encoder, s.feedService).Routes)
			r.Route("/irc", newIrcHandler(encoder, s.sse, s.ircService).Routes)
			r.Route("/indexer", newIndexerHandler(encoder, s.indexerService, s.ircService).Routes)
			r.Route("/lists", newListHandler(encoder, s.listService).Routes)
			r.Route("/logs", newLogsHandler(s.config).Routes)
			r.Route("/notification", newNotificationHandler(encoder, s.notificationService).Routes)
			r.Route("/proxy", newProxyHandler(encoder, s.proxyService).Routes)
			r.Route("/release", newReleaseHandler(encoder, s.releaseService).Routes)
			r.Route("/webhook", newWebhookHandler(encoder, s.listService).Routes)

			r.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
				s.sse.Headers = map[string]string{
					"Content-Type":      "text/event-stream",
					"Cache-Control":     "no-cache",
					"Connection":        "keep-alive",
					"X-Accel-Buffering": "no",
				}

				s.sse.ServeHTTP(w, r)
			})
		})
	})

	webRouter := chi.NewRouter()
	newWebHandler(s.log, web.DistDirFS).RegisterRoutes(webRouter)

	r.Mount("/", webRouter)
	r.Mount("/api", apiRouter)

	return r
}
