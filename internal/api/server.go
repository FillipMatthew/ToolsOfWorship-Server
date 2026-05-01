package api

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"sync"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	Logger        *slog.Logger
	config        config.ServerConfig
	healthChecker HealthChecker
	middleware    []MiddlewareFunc
	router        Router
	httpServer    *http.Server
	once          sync.Once
	cancel        func()
}

func NewServer(logger *slog.Logger, config config.ServerConfig, healthChecker HealthChecker, db *sql.DB, mw []MiddlewareFunc, rt Router) *Server {
	collector := NewDBStatsCollector(db)
	prometheus.MustRegister(collector)

	return &Server{
		Logger:        logger,
		config:        config,
		healthChecker: healthChecker,
		middleware:    mw,
		router:        rt,
	}
}

func (s *Server) Start(ctx context.Context) error {
	ctx, s.cancel = context.WithCancel(ctx)

	s.Logger.Info("starting server", slog.String("address", s.config.GetListenAddress()))

	mux := http.NewServeMux()
	s.setupHandlers(mux)

	s.httpServer = &http.Server{
		Addr:    s.config.GetListenAddress(),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		if err := s.Stop(); err != nil {
			s.Logger.Error("server stopped with a failure", "error", err)
		}
	}()

	s.Logger.Info("starting listener")

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	s.cancel()
	var err error
	s.once.Do(func() {
		s.Logger.Info("shutting down server")
		err = s.httpServer.Shutdown(context.Background())
	})

	return err
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	hh, err := s.healthChecker.CheckHealth(r.Context())
	if err != nil {
		RespondError(w, &Error{Code: http.StatusInternalServerError, Message: err.Error(), Err: err})
		return
	}

	RespondJSON(w, HealthResponse{Health: hh}, http.StatusOK)
}

func (s *Server) setupHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/health", s.health)
	mux.Handle("/metrics", promhttp.Handler())

	allMiddlewares := []MiddlewareFunc{
		WithTimeout(s.config.GetRequestTimeout()),
		SecurityHeadersMiddleware,
		CORSMiddleware(s.config.GetCORSAllowedOrigins()),
		WithLog(s.Logger),
		MetricsMiddleware,
		WithHTTPErrStatus,
	}
	allMiddlewares = append(allMiddlewares, s.middleware...)

	middlewareFunc := ChainMiddleware(allMiddlewares...)

	for _, rt := range s.router.Routes() {
		handler := middlewareFunc(rt.Method, rt.Pattern, rt.Handler)
		mux.Handle(rt.Pattern, DiscardError(rt.Method, handler))
	}
}
