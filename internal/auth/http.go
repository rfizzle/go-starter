package auth

import (
	"context"
	"net/http"
)

type Authorization func(ctx context.Context, r *http.Request) (context.Context, error)
