package web

import (
	"context"
	"net/http"
)

type Endpoint func(ctx context.Context, request interface{}, w http.ResponseWriter) (*HTTPSuccess, error)
