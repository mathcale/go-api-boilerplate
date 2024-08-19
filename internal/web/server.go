package web

import (
	"fmt"
	"net/http"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type Server interface {
	Start() error
}

type server struct {
	logger   logger.Logger
	router   *http.ServeMux
	handlers []handler
	port     int
}

func NewServer(l logger.Logger, handlers []handler, port int) Server {
	return &server{
		logger:   l,
		router:   http.NewServeMux(),
		handlers: handlers,
		port:     port,
	}
}

func (s *server) Start() error {
	for _, h := range s.handlers {
		s.logger.Debug("Registering route", map[string]interface{}{
			"method": h.method,
			"path":   h.path,
		})

		s.router.HandleFunc(fmt.Sprintf("%s %s", h.method, h.path), h.handlerFunc)
	}

	s.logger.Info("Starting server", map[string]interface{}{
		"port": s.port,
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.router)
}
