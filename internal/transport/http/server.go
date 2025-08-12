package http

import (
	"context"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type HTTPServer struct {
	server *http.Server
	log    *logrus.Logger
}

func NewServer(handler http.Handler, port string, log *logrus.Logger) *HTTPServer {
	s := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &HTTPServer{
		server: s,
		log:    log,
	}
}

func (s *HTTPServer) Run() error {
	s.log.Infof("Starting HTTP server on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.log.Info("Shutting down HTTP server gracefully...")
	return s.server.Shutdown(ctx)
}
