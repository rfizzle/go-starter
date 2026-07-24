package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

const loginPath = "/login"

type requestLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

// NewLoginRateLimit constructs middleware that limits only login requests.
// Forwarded client addresses are considered only when the immediate peer and
// every proxy to its right belong to trustedProxyCIDRs.
func NewLoginRateLimit(
	limiter requestLimiter,
	trustedProxyCIDRs []string,
	window time.Duration,
	logger *zap.Logger,
) (Constructor, error) {
	if limiter == nil {
		return nil, errors.New("login rate limiter is required")
	}
	if window <= 0 {
		return nil, errors.New("login rate limit window must be positive")
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	trustedProxies := make([]netip.Prefix, 0, len(trustedProxyCIDRs))
	for index, rawPrefix := range trustedProxyCIDRs {
		prefix, err := netip.ParsePrefix(strings.TrimSpace(rawPrefix))
		if err != nil || prefix != prefix.Masked() {
			return nil, fmt.Errorf("trusted proxy %d must be a canonical IP network in CIDR notation", index)
		}
		trustedProxies = append(trustedProxies, prefix)
	}

	retryAfter := strconv.FormatInt(int64((window+time.Second-1)/time.Second), 10)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != loginPath {
				next.ServeHTTP(w, r)
				return
			}

			clientIP, err := trustedClientIP(r, trustedProxies)
			if err != nil {
				logger.Error("derive login client identity", zap.Error(err))
				writeRateLimitError(w, http.StatusInternalServerError, "")
				return
			}

			allowed, err := limiter.Allow(r.Context(), clientIP.String())
			if err != nil {
				logger.Error("check login rate limit", zap.Stringer("client_ip", clientIP), zap.Error(err))
				writeRateLimitError(w, http.StatusInternalServerError, "")
				return
			}
			if !allowed {
				logger.Warn("login rate limit exceeded", zap.Stringer("client_ip", clientIP))
				writeRateLimitError(w, http.StatusTooManyRequests, retryAfter)
				return
			}

			next.ServeHTTP(w, r)
		})
	}, nil
}

func trustedClientIP(r *http.Request, trustedProxies []netip.Prefix) (netip.Addr, error) {
	peer, err := remoteIP(r.RemoteAddr)
	if err != nil {
		return netip.Addr{}, err
	}
	if !containsIP(trustedProxies, peer) {
		return peer, nil
	}

	forwardedValues := r.Header.Values("X-Forwarded-For")
	if len(forwardedValues) == 0 {
		return peer, nil
	}

	current := peer
	hops := strings.Split(strings.Join(forwardedValues, ","), ",")
	for index := len(hops) - 1; index >= 0; index-- {
		if !containsIP(trustedProxies, current) {
			break
		}

		hop, err := netip.ParseAddr(strings.TrimSpace(hops[index]))
		if err != nil {
			// A malformed forwarding chain is not a trustworthy source of
			// identity. Collapse it to the immediate peer rather than letting
			// attacker-controlled text create arbitrary limiter keys.
			return peer, nil
		}
		current = hop.Unmap()
	}

	return current, nil
}

func remoteIP(remoteAddr string) (netip.Addr, error) {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}

	address, err := netip.ParseAddr(strings.TrimSpace(host))
	if err != nil {
		return netip.Addr{}, fmt.Errorf("parse request peer address: %w", err)
	}
	return address.Unmap(), nil
}

func containsIP(prefixes []netip.Prefix, address netip.Addr) bool {
	for _, prefix := range prefixes {
		if prefix.Contains(address) {
			return true
		}
	}
	return false
}

func writeRateLimitError(w http.ResponseWriter, statusCode int, retryAfter string) {
	w.Header().Set("Content-Type", "application/json")
	if retryAfter != "" {
		w.Header().Set("Retry-After", retryAfter)
	}
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	}{
		Code:    statusCode,
		Message: http.StatusText(statusCode),
		Status:  "fail",
	})
}
