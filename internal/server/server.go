package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rfizzle/go-starter/internal/controller"

	"go.uber.org/zap"

	"github.com/rfizzle/go-starter/internal/server/middleware"
)

type Server struct {
	httpServer   *http.Server
	fileServer   http.Handler
	logger       *zap.Logger
	address      string
	port         int
	writeTimeout int
	readTimeout  int
	idleTimeout  int
}

func New(controller *controller.Controller, options ...Option) (*Server, error) {
	// Initialize server with default values
	server := defaultServer()

	// Apply options
	for _, v := range options {
		v(server)
	}

	// Set up API
	apiHandler, err := setupApi(server.logger, controller)
	if err != nil {
		return nil, err
	}

	// Set up middleware
	handlerWithMiddleware := middleware.New(server.logger, apiHandler, server.fileServer)

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
	if err := s.httpServer.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		} else {
			return fmt.Errorf("issue with httpServer.ListenAndServe(): %w", err)
		}
	}
	return nil
}

func (s *Server) Stop() error {
	// Setup context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Shutdown server
	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("server.Shutdown(ctx): %w", err)
	}

	return nil
}

func defaultServer() *Server {
	return &Server{
		httpServer:   nil,
		fileServer:   defaultFileServer(),
		logger:       zap.NewNop(),
		address:      "0.0.0.0",
		port:         8080,
		readTimeout:  10,
		writeTimeout: 10,
		idleTimeout:  10,
	}
}

func defaultFileServer() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
}
