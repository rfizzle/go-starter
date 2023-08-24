package middleware

import (
  "net/http"
  "time"

  "go.uber.org/zap"

  "github.com/go-chi/chi/v5/middleware"
)

func New(logger *zap.Logger, apiHandler, staticHandler http.Handler) http.Handler {
  return NewChain(
    middleware.Timeout(60*time.Second),
    middleware.RequestID,
    middleware.RealIP,
    zapLogger(logger),
    middleware.Recoverer,
    SwaggerUI(),
    RedocUI(),
    fileServer(staticHandler),
  ).Then(apiHandler)
}
