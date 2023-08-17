package cli

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/rfizzle/go-starter/internal/app"
)

// GracefulExit creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS.
func GracefulExit(app *app.App) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		app.Logger().Info("gracefully shutting down...")

		// PreStopHook will set app readiness to false (for kubernetes)
		app.SetReady(false)

		// Sleep to let kubernetes state get synced
		app.Logger().Debug("20 second sleep for kubernetes state sync...")
		time.Sleep(time.Second * 20)

		// Execute safe stop function
		err := app.Stop()
		if err != nil {
			app.Logger().Error("error stopping application", zap.Error(err))
		}
	}()
}
