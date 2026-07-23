package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// panicsTotal counts the panics recovered by the Recoverer middleware. It is
// registered on the default Prometheus registry so it is exposed by the same
// /metrics handler as the rest of the server's metrics.
var panicsTotal = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "http_panics_recovered_total",
	Help: "Total number of panics recovered by the HTTP panic-recovery middleware.",
})

func init() {
	prometheus.MustRegister(panicsTotal)
}

// Recoverer is a panic-recovery middleware that logs the panic to the provided
// zap logger (with the request ID and a stack trace), increments a panic
// counter, and responds with 500 Internal Server Error.
//
// It exists to replace chi's built-in Recoverer, which writes to the stdlib
// logger and is therefore invisible in the structured JSON logs this server
// emits. As with chi's version, http.ErrAbortHandler is not recovered so the
// connection is aborted without a spurious log entry.
func Recoverer(logger *zap.Logger) func(next http.Handler) http.Handler {
	if logger == nil {
		logger = zap.NewNop()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Wrap the writer so we can tell whether the handler already sent a
			// status before panicking. Without this, a panic mid-stream would
			// trigger a superfluous WriteHeader call — which the stdlib logs to
			// the very logger this middleware exists to bypass.
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				rvr := recover()
				if rvr == nil {
					return
				}

				// Let http.ErrAbortHandler propagate so the server aborts the
				// response as intended; it is not a real panic to report.
				if rvr == http.ErrAbortHandler { //nolint:errorlint // sentinel is compared by identity, matching net/http
					panic(rvr)
				}

				panicsTotal.Inc()

				fields := []zap.Field{
					zap.Any("panic", rvr),
					zap.String("method", r.Method),
					zap.String("request", r.RequestURI),
					zap.String("remote", r.RemoteAddr),
					zap.ByteString("stack", debug.Stack()),
				}
				if requestID := middleware.GetReqID(r.Context()); requestID != "" {
					fields = append(fields, zap.String("request-id", requestID))
				}
				logger.Error("recovered from panic", fields...)

				// Only send a 500 if the handler hasn't already written a status
				// (ww.Status() == 0) and the connection isn't being upgraded.
				// Writing again on an already-committed response is a no-op that
				// the stdlib complains about.
				if ww.Status() == 0 && r.Header.Get("Connection") != "Upgrade" {
					ww.WriteHeader(http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
