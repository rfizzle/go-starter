package server

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/rfizzle/go-starter/internal/server/middleware"
)

type Server struct {
	httpServer   *http.Server
	logger       *zap.Logger
	address      string
	port         int
	writeTimeout int
	readTimeout  int
	idleTimeout  int
}

func New(options ...Option) (*Server, error) {
	// Initialize server with default values
	server := defaultServer()

	// Apply options
	for _, v := range options {
		v(server)
	}

	// Set up API
	apiHandler, err := setupApi(server.logger, nil)
	if err != nil {
		return nil, err
	}

	// Set up middleware
	handlerWithMiddleware := middleware.New(server.logger, apiHandler)

	// Set up http server
	httpServer := &http.Server{
		Addr: fmt.Sprintf("%s:%d", server.address, server.port),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * time.Duration(server.writeTimeout),
		ReadTimeout:  time.Second * time.Duration(server.readTimeout),
		IdleTimeout:  time.Second * time.Duration(server.idleTimeout),
		Handler:      handlerWithMiddleware,
	}

	// Set the http server on the server
	server.httpServer = httpServer

	return server, nil
}

func (s *Server) Start() error {
	return nil
}

func (s *Server) Stop() error {
	return nil
}

func defaultServer() *Server {
	return &Server{
		httpServer:   nil,
		logger:       zap.NewNop(),
		address:      "0.0.0.0",
		port:         8080,
		readTimeout:  10,
		writeTimeout: 10,
		idleTimeout:  10,
	}
}
