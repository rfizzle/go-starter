//go:build integration

package cmd

import (
	"net"
	"os"
	"strings"
	"testing"

	"github.com/rfizzle/go-starter/internal/app"
	"github.com/rfizzle/go-starter/internal/platform/config"
	"github.com/rfizzle/go-starter/internal/platform/database/postgrestest"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	os.Exit(postgrestest.RunMain(m))
}

func TestServeCommand_RealApplicationStartupFailureUsesCobraErrorPath(t *testing.T) {
	database := postgrestest.Shared(t).Database(t)

	blocker, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("bind occupied listener: %v", err)
	}
	defer func() { _ = blocker.Close() }()

	address, ok := blocker.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("listener address type = %T, want *net.TCPAddr", blocker.Addr())
	}

	cfg := validServeCommandConfig()
	cfg.Webserver.Address = "127.0.0.1"
	cfg.Webserver.Port = address.Port
	cfg.Webserver.ShutdownTimeout = 5
	cfg.Database.URI = database.URI
	cfg.Database.PingTimeout = 5
	cfg.Database.CloseTimeout = 5

	deps := serveDependencies{
		resolveConfig: func() (*config.Application, error) {
			return cfg, nil
		},
		newLogger: func(*config.Application) (*zap.Logger, error) {
			return zap.NewNop(), nil
		},
		newApplication: func(cfg *config.Application) (serveApplication, error) {
			return app.New(cfg)
		},
		syncLogger: func(*zap.Logger) error {
			return nil
		},
	}

	err = newServeCommand(deps).ExecuteContext(t.Context())
	if err == nil {
		t.Fatal("ExecuteContext() error = nil, want listener startup failure")
	}
	if !strings.Contains(err.Error(), "run application") {
		t.Errorf("ExecuteContext() error = %q, want Cobra run context", err)
	}
	if !strings.Contains(err.Error(), "issue opening webserver listener") {
		t.Errorf("ExecuteContext() error = %q, want listener startup context", err)
	}
}
