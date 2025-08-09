package api

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"sync"

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

	s.Logger.Println("TLS:", s.config.IsTLS())

	if s.config.IsTLS() {
		s.Logger.Println("cert:", s.config.GetCertPath())
		s.Logger.Println("key:", s.config.GetKeyPath())

		cfg := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		}

		s.httpServer = &http.Server{
			Addr:         s.config.GetListenAddress(),
			Handler:      mux,
			TLSConfig:    cfg,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		}

		go func() {
			<-ctx.Done()
			if err := s.Stop(); err != nil {
				s.Logger.Fatalf("server stopped with a failure: %v", err)
			}
		}()

		s.Logger.Printf("starting listener")

		if err := s.httpServer.ListenAndServeTLS(s.config.GetCertPath(), s.config.GetKeyPath()); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	} else {
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
	mux.Handle("/", http.FileServer(http.Dir(config.GetPublicDir())))

	allMiddlewares := append(s.middleware, WithLog(s.Logger))
	allMiddlewares = append(allMiddlewares, WithHTTPErrStatus)

	middlewareFunc := ChainMiddleware(
		allMiddlewares...,
	)

	for _, rt := range s.router.Routes() {
		handler := middlewareFunc(rt.Method, rt.Pattern, rt.Handler)
		mux.Handle(rt.Pattern, DiscardError(rt.Method, handler))
	}
}
