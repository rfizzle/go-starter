package templateinit

import (
	"bytes"
	"context"
	"errors"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitialize(t *testing.T) {
	root := newTemplateRepository(t)
	gitSentinel := filepath.Join(root, ".git", "initializer-sentinel")
	if err := os.WriteFile(gitSentinel, []byte(defaultModulePath), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", gitSentinel, err)
	}

	cfg := Config{
		Root:            root,
		ModulePath:      "example.com/acme/orders",
		ApplicationName: "Acme Orders",
		BinaryName:      "orders-api",
		EnvPrefix:       "ORDERS",
		ImageName:       "registry.example.com/acme/orders-api",
		Description:     "Order processing API",
	}
	if err := Initialize(context.Background(), cfg); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	assertFileContains(t, filepath.Join(root, "go.mod"), "module "+cfg.ModulePath)
	assertFileContains(t, filepath.Join(root, "cmd", "root.go"), `Use:   "`+cfg.BinaryName+`"`)
	assertFileContains(t, filepath.Join(root, "cmd", "root.go"), `Short: "`+cfg.Description+`"`)
	assertFileContains(t, filepath.Join(root, "cmd", "root.go"), `SetEnvPrefix("`+cfg.EnvPrefix+`")`)
	assertFileContains(t, filepath.Join(root, ".goreleaser.yaml"), "project_name: "+cfg.BinaryName)
	assertFileContains(t, filepath.Join(root, ".goreleaser.yaml"), "- "+cfg.ImageName)
	assertFileContains(t, filepath.Join(root, "api", "openapi.yml"), `title: "`+cfg.ApplicationName+`"`)
	assertFileContains(t, filepath.Join(root, "api", "openapi.yml"), `description: "`+cfg.Description+`"`)
	assertFileContains(t, filepath.Join(root, "web", "build", "index.html"), cfg.ApplicationName)
	for _, removed := range removedPaths {
		if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(removed))); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("Lstat(%q) error = %v, want not exist after initialization", removed, err)
		}
	}

	assertFileEquals(t, gitSentinel, defaultModulePath)
	assertFileEquals(t, filepath.Join(root, "LICENSE"), "historical attribution: "+defaultModulePath+"\n")
	assertFileEquals(t, filepath.Join(root, "internal", "api", "v1", "oas_cfg_gen.go"), "package v1\n")
	assertFileEquals(t, filepath.Join(root, "internal", "platform", "database", "postgres", "db", "db.go"), "package db\n")

	for _, path := range trackedPaths(t, root) {
		if skipTrackedPath(filepath.ToSlash(path)) {
			continue
		}
		contents, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", path, err)
		}
		if filepath.Ext(path) == ".go" {
			formatted, err := format.Source(contents)
			if err != nil {
				t.Errorf("format.Source(%s) error = %v", path, err)
			} else if !bytes.Equal(contents, formatted) {
				t.Errorf("%s is not gofmt-formatted after initialization", path)
			}
		}
		for _, obsolete := range []string{
			defaultModulePath,
			defaultImageName,
			defaultApplicationName,
			defaultEnvPrefix,
			defaultBinaryName,
			defaultProjectSlug,
		} {
			if strings.Contains(string(contents), obsolete) {
				t.Errorf("%s still contains obsolete identifier %q", path, obsolete)
			}
		}
	}

	// The initializer sources are gone, so a second run is only possible from a
	// copy of the tool; the rewritten go.mod must still make it refuse.
	if err := Initialize(context.Background(), cfg); err == nil {
		t.Fatal("second Initialize() error = nil, want template check failure")
	}
}

func TestInitializeRefusesDirtyWorktreeAndPreservesUntrackedFiles(t *testing.T) {
	root := newTemplateRepository(t)
	untrackedPath := filepath.Join(root, "personal-notes.txt")
	contents := "do not rewrite " + defaultModulePath
	if err := os.WriteFile(untrackedPath, []byte(contents), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg := Config{
		Root:            root,
		ModulePath:      "example.com/acme/orders",
		ApplicationName: "Acme Orders",
		BinaryName:      "orders-api",
		EnvPrefix:       "ORDERS",
		ImageName:       "registry.example.com/acme/orders-api",
	}
	if err := Initialize(context.Background(), cfg); !errors.Is(err, ErrDirtyWorktree) {
		t.Fatalf("Initialize() error = %v, want ErrDirtyWorktree", err)
	}

	assertFileEquals(t, untrackedPath, contents)
	assertFileContains(t, filepath.Join(root, "go.mod"), defaultModulePath)
	for _, removed := range removedPaths {
		if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(removed))); err != nil {
			t.Errorf("Lstat(%q) error = %v, want initializer sources preserved", removed, err)
		}
	}
}

// TestInitializeWithBinaryNameContainingTemplateToken guards the rewrite against
// cascading substitution: the substituted names below all contain "starter", so a
// replacement pass that rescans its own output doubles the prefix (mymystarter).
// Initialization cannot be repeated, which makes that corruption unrecoverable.
func TestInitializeWithBinaryNameContainingTemplateToken(t *testing.T) {
	root := newTemplateRepository(t)
	cfg := Config{
		Root:            root,
		ModulePath:      "example.com/acme/mystarter",
		ApplicationName: "My Starter",
		BinaryName:      "mystarter",
		EnvPrefix:       "MYSTARTER",
		ImageName:       "registry.example.com/acme/mystarter",
		Description:     "My starter service",
	}
	if err := Initialize(context.Background(), cfg); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	for _, path := range trackedPaths(t, root) {
		if skipTrackedPath(filepath.ToSlash(path)) {
			continue
		}
		contents, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", path, err)
		}
		if strings.Contains(string(contents), "my"+cfg.BinaryName) {
			t.Errorf("%s = %q, want no cascaded binary name", path, contents)
		}
	}

	assertFileContains(t, filepath.Join(root, "Makefile"), `APP_NAME="`+cfg.BinaryName+`"`)
	assertFileContains(t, filepath.Join(root, "Makefile"), `APP_PACKAGE="`+cfg.ModulePath+`"`)
	assertFileContains(t, filepath.Join(root, ".goreleaser.yaml"), "project_name: "+cfg.BinaryName)
	assertFileContains(t, filepath.Join(root, "build", "docker", "Dockerfile"), "-o /out/"+cfg.BinaryName+" .")
	assertFileContains(t, filepath.Join(root, "build", "docker", "Dockerfile.release"), "${TARGETPLATFORM}/"+cfg.BinaryName+" /app")
	assertFileContains(t, filepath.Join(root, "scripts", "container-build.bash"), `APPLICATION_NAME="${APPLICATION_NAME:-`+cfg.BinaryName+`}"`)
}

func TestNormalizeAndValidate(t *testing.T) {
	valid := Config{
		ModulePath:      "example.com/acme/orders/v2",
		ApplicationName: "Acme Orders",
		BinaryName:      "orders-api",
		EnvPrefix:       "ORDERS_API",
		ImageName:       "localhost:5000/acme/orders-api",
	}

	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{name: "missing module", mutate: func(cfg *Config) { cfg.ModulePath = "" }},
		{name: "module with scheme", mutate: func(cfg *Config) { cfg.ModulePath = "https://example.com/acme/orders" }},
		{name: "module without path element", mutate: func(cfg *Config) { cfg.ModulePath = "example.com" }},
		{name: "module with dot element", mutate: func(cfg *Config) { cfg.ModulePath = "example.com/acme/../orders" }},
		{name: "multiline application", mutate: func(cfg *Config) { cfg.ApplicationName = "Acme\nOrders" }},
		{name: "uppercase binary", mutate: func(cfg *Config) { cfg.BinaryName = "Orders" }},
		{name: "lowercase environment prefix", mutate: func(cfg *Config) { cfg.EnvPrefix = "orders" }},
		{name: "tagged image", mutate: func(cfg *Config) { cfg.ImageName = "registry.example.com/acme/orders:v1" }},
		{name: "multiline description", mutate: func(cfg *Config) { cfg.Description = "Orders\nAPI" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := valid
			tt.mutate(&cfg)
			if _, err := normalizeAndValidate(cfg); err == nil {
				t.Fatal("normalizeAndValidate() error = nil, want error")
			}
		})
	}

	got, err := normalizeAndValidate(valid)
	if err != nil {
		t.Fatalf("normalizeAndValidate() error = %v", err)
	}
	if got.Description != valid.ApplicationName+" service" {
		t.Errorf("Description = %q, want default based on application name", got.Description)
	}
}

func newTemplateRepository(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	files := map[string]string{
		"go.mod": "module " + defaultModulePath + "\n\ngo 1.26.5\n",
		"Makefile": `APP_NAME="gostarter"
APP_PACKAGE="github.com/rfizzle/go-starter"
`,
		".goreleaser.yaml": `project_name: starter
builds:
  - id: starter
    binary: starter
dockers_v2:
  - images:
      - ghcr.io/rfizzle/go-starter
    labels:
      org.opencontainers.image.description: OpenAPI-first foundation for PostgreSQL-backed Go services
`,
		"api/openapi.yml": `openapi: 3.1.0
info:
  title: Go Starter Project
`,
		"build/docker/Dockerfile": `RUN go build -ldflags="-X github.com/rfizzle/go-starter/cmd.ApplicationName=starter" -o /out/starter .
LABEL org.opencontainers.image.description="OpenAPI-first foundation for PostgreSQL-backed Go services" \
    org.opencontainers.image.source="https://github.com/rfizzle/go-starter" \
    io.github.rfizzle.go-starter.environment="production"
COPY --from=build /out/starter /app
`,
		"build/docker/Dockerfile.release": "COPY ${TARGETPLATFORM}/starter /app\n",
		"cmd/root.go": `package cmd
var root = struct {
	Use string
	Short string
}{
	Use: "go-starter",
	Short:   "A starter golang project",
}
func configure() { SetEnvPrefix("GOSTARTER") }
`,
		"main.go": `package main
import "github.com/rfizzle/go-starter/cmd"
func main() { cmd.Execute() }
`,
		"internal/modules/user/repository_test.go": `package user
import (
	"github.com/jackc/pgx/v5"
	"github.com/rfizzle/go-starter/internal/platform/database"
	"github.com/stretchr/testify/assert"
)
var _ = pgx.ErrNoRows
var _ = database.StatusNone
var _ = assert.Equal
`,
		"README.md": `# Go Starter
Go Starter is a starter for services.
Run ./bin/gostarter or use ghcr.io/rfizzle/go-starter.
`,
		"LICENSE":                        "historical attribution: " + defaultModulePath + "\n",
		"internal/api/v1/oas_cfg_gen.go": "package v1\n",
		"internal/templateinit/initializer.go": `package templateinit
const defaultModulePath = "github.com/rfizzle/go-starter"
`,
		"tools/initialize/main.go": `package main
import "github.com/rfizzle/go-starter/internal/templateinit"
`,
		"internal/platform/database/postgres/db/db.go":          "package db\n",
		"scripts/container-build.bash":                          `APPLICATION_NAME="${APPLICATION_NAME:-starter}"` + "\n",
		"docs/development/environment.md":                       "Use GOSTARTER_DATABASE_URI for gostarter.\n",
		"docs/development/ogen-workflow.md":                     "Import github.com/rfizzle/go-starter/internal/api/v1.\n",
		"internal/platform/observability/observability.go":      `const serviceName = "go-starter"` + "\n",
		"internal/platform/observability/observability_test.go": `const wantServiceName = "go-starter"` + "\n",
		"web/build/index.html":                                  "<h1>Go Starter</h1>\n",
	}

	for path, contents := range files {
		fullPath := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("MkdirAll(%q) error = %v", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(contents), 0o600); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", path, err)
		}
	}

	runGit(t, root, "init", "-q")
	runGit(t, root, "add", ".")
	runGit(t, root, "-c", "user.name=Template Test", "-c", "user.email=template@example.com", "commit", "-qm", "fixture")
	return root
}

func trackedPaths(t *testing.T, root string) []string {
	t.Helper()

	command := exec.Command("git", "-c", "core.fsmonitor=false", "-C", root, "ls-files")
	out, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git ls-files error = %v: %s", err, out)
	}
	return strings.Fields(string(out))
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()

	commandArgs := append([]string{"-c", "core.fsmonitor=false", "-C", root}, args...)
	command := exec.Command("git", commandArgs...)
	if out, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %s error = %v: %s", strings.Join(args, " "), err, out)
	}
}

func assertFileContains(t *testing.T, path, want string) {
	t.Helper()

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if !strings.Contains(string(contents), want) {
		t.Errorf("%s = %q, want substring %q", path, contents, want)
	}
}

func assertFileEquals(t *testing.T, path, want string) {
	t.Helper()

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if string(contents) != want {
		t.Errorf("%s = %q, want %q", path, contents, want)
	}
}
