package cli

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/rfizzle/go-starter/internal/app"
)

// WithGracefulShutdown derives an application context from parent (typically a
// signal.NotifyContext). When parent is cancelled — e.g. on SIGINT/SIGTERM — it
// marks the app not-ready and, after the configured drain delay, cancels the
// returned context so the app's components begin graceful shutdown.
//
// The drain delay gives orchestrators such as Kubernetes time to observe the
// failing readiness probe and stop routing traffic before components tear down.
// A delay of zero skips the wait. Invoking the returned CancelFunc also releases
// the goroutine, so the helper does not leak when the app exits on its own
// (e.g. a component fails before any signal arrives). The CancelFunc must be
// called to release resources.
func WithGracefulShutdown(parent context.Context, application *app.App, delay time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		select {
		case <-parent.Done():
			// Signal-driven shutdown: fall through to the drain sequence below.
		case <-ctx.Done():
			// The app already stopped and the caller cancelled us; nothing to drain.
			return
		}

		application.Logger().Info("gracefully shutting down...")

		// Fail readiness probes before draining so traffic is routed away first.
		application.SetReady(false)

		if delay > 0 {
			application.Logger().Info("waiting out shutdown delay before draining", zap.Duration("delay", delay))
			time.Sleep(delay)
		}

		cancel()
	}()

	return ctx, cancel
}
