package api

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api/middleware"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/config"
)

type Server struct {
	Logger        *log.Logger
	config        config.ServerConfig
	healthChecker HealthChecker
	middleware    []MiddlewareFunc
	router        Router
	httpServer    *http.Server
	once          sync.Once
	cancel        func()
}

func NewServer(logger *log.Logger, config config.ServerConfig, healthChecker HealthChecker, mw []MiddlewareFunc, rt Router) *Server {
	server := &Server{
		Logger:        logger,
		config:        config,
		healthChecker: healthChecker,
		middleware:    mw,
		router:        rt,
	}

	return server
}

func (s *Server) Start(ctx context.Context) error {
	ctx, s.cancel = context.WithCancel(ctx)

	s.Logger.Println("starting server:")
	s.Logger.Printf("address: '%s'\n", s.config.GetListenAddress())

	mux := http.NewServeMux()
	s.setupHandlers(s.config, mux)

	s.httpServer = &http.Server{
		Addr:    s.config.GetListenAddress(),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		if err := s.Stop(); err != nil {
			s.Logger.Fatalf("server stopped with a failure: %v", err)
		}
	}()

	s.Logger.Printf("starting listener")

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	s.cancel()
	var err error
	s.once.Do(func() {
		s.Logger.Println("shutting down server")
		err = s.httpServer.Shutdown(context.Background())
	})

	return err
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	hh, err := s.healthChecker.CheckHealth(r.Context())
	if err != nil {
		//w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		RespondError(w, &Error{Code: http.StatusInternalServerError, Message: err.Error(), Err: err})
		return
	}

	RespondJSON(w, HealthResponse{Health: hh}, http.StatusOK)
}

func (s *Server) setupHandlers(config config.ServerConfig, mux *http.ServeMux) {
	mux.HandleFunc("/health", s.health)

	allMiddlewares := []MiddlewareFunc{
		middleware.SecurityHeadersMiddleware,
		middleware.CORSMiddleware,
		WithLog(s.Logger),
		WithHTTPErrStatus,
	}
	allMiddlewares = append(allMiddlewares, s.middleware...)

	middlewareFunc := ChainMiddleware(
		allMiddlewares...,
	)

	for _, rt := range s.router.Routes() {
		handler := middlewareFunc(rt.Method, rt.Pattern, rt.Handler)
		mux.Handle(rt.Pattern, DiscardError(rt.Method, handler))
	}
}
