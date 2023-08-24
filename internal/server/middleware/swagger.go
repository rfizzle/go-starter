package middleware

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

func SwaggerUI() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		opts := middleware.SwaggerUIOpts{}
		opts.EnsureDefaults()
		opts.BasePath = "/api/docs"
		opts.Path = "swagger"
		return middleware.SwaggerUI(opts, next)
	}
}

func RedocUI() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		opts := middleware.RedocOpts{}
		opts.EnsureDefaults()
		opts.BasePath = "/api/docs"
		opts.Path = "redoc"
		return middleware.Redoc(opts, next)
	}
}
