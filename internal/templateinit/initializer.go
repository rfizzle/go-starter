// Package templateinit safely replaces the template's project identity.
//
// Initialize is a one-shot operation: after rewriting the tracked files it
// deletes this package and its command entrypoint from the repository, so a
// derived project carries no initializer code and cannot run it twice.
package templateinit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	defaultApplicationName = "Go Starter"
	defaultBinaryName      = "gostarter"
	defaultEnvPrefix       = "GOSTARTER"
	defaultImageName       = "ghcr.io/rfizzle/go-starter"
	defaultModulePath      = "github.com/rfizzle/go-starter"
	defaultProjectSlug     = "go-starter"
	defaultReleaseName     = "starter"
)

// ErrDirtyWorktree reports that the repository has staged, unstaged, or
// untracked changes.
var ErrDirtyWorktree = errors.New("worktree is not clean")

var (
	binaryNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)
	envPrefixPattern  = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	imageNamePattern  = regexp.MustCompile(`^[a-z0-9]+(?:[._-][a-z0-9]+)*(?::[0-9]+)?(?:/[a-z0-9]+(?:[._-][a-z0-9]+)*)+$`)
	// modulePathPattern requires a lowercase dotted host followed by at least
	// one path element; elements may not start or end with a dot.
	modulePathPattern = regexp.MustCompile(`^[a-z0-9-]+(?:\.[a-z0-9-]+)+(?:/[A-Za-z0-9_~-](?:[A-Za-z0-9._~-]*[A-Za-z0-9_~-])?)+$`)
)

// removedPaths are the one-shot initializer sources that Initialize deletes
// as its final step. Their absence is what prevents a second run.
var removedPaths = []string{
	"internal/templateinit",
	"tools/initialize",
}

// Config describes the new identity and the repository to initialize.
type Config struct {
	Root            string
	ModulePath      string
	ApplicationName string
	BinaryName      string
	EnvPrefix       string
	ImageName       string
	Description     string
}

type fileChange struct {
	path     string
	original []byte
	updated  []byte
	mode     os.FileMode
}

// Initialize validates cfg, verifies that Root is the clean template repository,
// replaces project identifiers in tracked, hand-written files, and finally
// deletes the initializer sources listed in removedPaths.
func Initialize(ctx context.Context, cfg Config) error {
	cfg, err := normalizeAndValidate(cfg)
	if err != nil {
		return err
	}

	root, err := repositoryRoot(ctx, cfg.Root)
	if err != nil {
		return err
	}
	cfg.Root = root

	if err := requireTemplate(root); err != nil {
		return err
	}
	if err := requireCleanWorktree(ctx, root); err != nil {
		return err
	}

	paths, err := trackedFiles(ctx, root)
	if err != nil {
		return err
	}

	changes, err := planChanges(root, paths, cfg)
	if err != nil {
		return err
	}
	if len(changes) == 0 {
		return errors.New("template identifiers were not found")
	}

	applied := make([]fileChange, 0, len(changes))
	for _, change := range changes {
		if err := writeFileAtomically(change.path, change.updated, change.mode); err != nil {
			return errors.Join(
				fmt.Errorf("updating %s: %w", change.path, err),
				rollback(applied),
			)
		}
		applied = append(applied, change)
	}

	for _, relativePath := range removedPaths {
		if err := os.RemoveAll(filepath.Join(root, relativePath)); err != nil {
			return errors.Join(
				fmt.Errorf("removing initializer sources %s: %w", relativePath, err),
				rollback(applied),
			)
		}
	}

	return nil
}

func normalizeAndValidate(cfg Config) (Config, error) {
	cfg.Root = strings.TrimSpace(cfg.Root)
	cfg.ModulePath = strings.TrimSpace(cfg.ModulePath)
	cfg.ApplicationName = strings.TrimSpace(cfg.ApplicationName)
	cfg.BinaryName = strings.TrimSpace(cfg.BinaryName)
	cfg.EnvPrefix = strings.TrimSpace(cfg.EnvPrefix)
	cfg.ImageName = strings.TrimSpace(cfg.ImageName)
	cfg.Description = strings.TrimSpace(cfg.Description)

	if cfg.Root == "" {
		cfg.Root = "."
	}
	if cfg.Description == "" && cfg.ApplicationName != "" {
		cfg.Description = cfg.ApplicationName + " service"
	}

	var errs error
	if err := validateModulePath(cfg.ModulePath); err != nil {
		errs = errors.Join(errs, err)
	}
	if err := validateDisplayText("application name", cfg.ApplicationName, 80); err != nil {
		errs = errors.Join(errs, err)
	}
	if !binaryNamePattern.MatchString(cfg.BinaryName) || len(cfg.BinaryName) > 63 {
		errs = errors.Join(errs, errors.New("binary name must be 1-63 lowercase letters, digits, or single hyphens and start with a letter"))
	}
	if !envPrefixPattern.MatchString(cfg.EnvPrefix) || len(cfg.EnvPrefix) > 63 {
		errs = errors.Join(errs, errors.New("environment prefix must be 1-63 uppercase letters, digits, or underscores and start with a letter"))
	}
	if !imageNamePattern.MatchString(cfg.ImageName) || len(cfg.ImageName) > 255 {
		errs = errors.Join(errs, errors.New("image name must be a lowercase, untagged OCI repository name such as ghcr.io/acme/orders"))
	}
	if err := validateDisplayText("description", cfg.Description, 200); err != nil {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return Config{}, fmt.Errorf("invalid initializer configuration: %w", errs)
	}
	return cfg, nil
}

func validateModulePath(modulePath string) error {
	if len(modulePath) > 255 || !modulePathPattern.MatchString(modulePath) {
		return errors.New("module path must be a lowercase dotted host followed by at least one path element, such as example.com/acme/orders")
	}
	return nil
}

func validateDisplayText(field, value string, maxLen int) error {
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if !utf8.ValidString(value) || len(value) > maxLen {
		return fmt.Errorf("%s must be valid UTF-8 no longer than %d bytes", field, maxLen)
	}
	if strings.ContainsAny(value, "\r\n\x00") {
		return fmt.Errorf("%s must be a single line without NUL bytes", field)
	}
	return nil
}

func repositoryRoot(ctx context.Context, requestedRoot string) (string, error) {
	absoluteRoot, err := filepath.Abs(requestedRoot)
	if err != nil {
		return "", fmt.Errorf("resolving repository path: %w", err)
	}
	absoluteRoot, err = filepath.EvalSymlinks(absoluteRoot)
	if err != nil {
		return "", fmt.Errorf("resolving repository symlinks: %w", err)
	}

	out, err := git(ctx, absoluteRoot, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("locating Git repository: %w", err)
	}
	gitRoot, err := filepath.EvalSymlinks(strings.TrimSpace(string(out)))
	if err != nil {
		return "", fmt.Errorf("resolving Git repository root: %w", err)
	}
	if gitRoot != absoluteRoot {
		return "", fmt.Errorf("initializer must run at repository root %s", gitRoot)
	}
	return gitRoot, nil
}

func requireTemplate(root string) error {
	goMod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return fmt.Errorf("reading go.mod: %w", err)
	}
	want := []byte("module " + defaultModulePath + "\n")
	if !bytes.HasPrefix(goMod, want) {
		return fmt.Errorf("repository is not the %s template at module %s", defaultApplicationName, defaultModulePath)
	}

	for _, path := range []string{
		"Makefile",
		".goreleaser.yaml",
		"api/openapi.yml",
		"build/docker/Dockerfile",
		"build/docker/Dockerfile.release",
		"cmd/root.go",
		"README.md",
	} {
		info, err := os.Lstat(filepath.Join(root, path))
		if err != nil {
			return fmt.Errorf("checking required template file %s: %w", path, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("required template file %s is not a regular file", path)
		}
	}
	return nil
}

func requireCleanWorktree(ctx context.Context, root string) error {
	out, err := git(ctx, root, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		return fmt.Errorf("checking worktree status: %w", err)
	}
	if len(out) != 0 {
		return fmt.Errorf("%w:\n%s", ErrDirtyWorktree, bytes.TrimSpace(out))
	}
	return nil
}

func trackedFiles(ctx context.Context, root string) ([]string, error) {
	out, err := git(ctx, root, "ls-files", "-z")
	if err != nil {
		return nil, fmt.Errorf("listing tracked files: %w", err)
	}

	rawPaths := bytes.Split(out, []byte{0})
	paths := make([]string, 0, len(rawPaths))
	for _, rawPath := range rawPaths {
		if len(rawPath) == 0 {
			continue
		}
		path := filepath.Clean(string(rawPath))
		if filepath.IsAbs(path) || path == "." || path == ".." || strings.HasPrefix(path, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("git returned unsafe tracked path %q", rawPath)
		}
		if skipTrackedPath(filepath.ToSlash(path)) {
			continue
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func skipTrackedPath(path string) bool {
	if path == "LICENSE" ||
		strings.HasPrefix(path, "internal/api/v1/") ||
		strings.HasPrefix(path, "internal/platform/database/postgres/db/") {
		return true
	}
	for _, removed := range removedPaths {
		if strings.HasPrefix(path, removed+"/") {
			return true
		}
	}
	return false
}

func planChanges(root string, paths []string, cfg Config) ([]fileChange, error) {
	changes := make([]fileChange, 0)
	for _, relativePath := range paths {
		path := filepath.Join(root, relativePath)
		info, err := os.Lstat(path)
		if err != nil {
			return nil, fmt.Errorf("checking tracked file %s: %w", relativePath, err)
		}
		if !info.Mode().IsRegular() {
			continue
		}

		original, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading tracked file %s: %w", relativePath, err)
		}
		if bytes.IndexByte(original, 0) >= 0 || !utf8.Valid(original) {
			continue
		}

		updated, err := rewrite(relativePath, original, cfg)
		if err != nil {
			return nil, fmt.Errorf("rewriting tracked file %s: %w", relativePath, err)
		}
		if bytes.Equal(original, updated) {
			continue
		}
		changes = append(changes, fileChange{
			path:     path,
			original: original,
			updated:  updated,
			mode:     info.Mode().Perm(),
		})
	}
	return changes, nil
}

type replacement struct {
	old string
	new string
}

func rewrite(path string, contents []byte, cfg Config) ([]byte, error) {
	text := string(contents)
	sourceURL := "https://" + cfg.ModulePath
	annotationKey := "io." + sanitizeAnnotation(cfg.ModulePath) + ".environment"

	replacements := []replacement{
		{defaultImageName, cfg.ImageName},
		{"https://" + defaultModulePath + ".git", sourceURL + ".git"},
		{"https://" + defaultModulePath, sourceURL},
		{defaultModulePath, cfg.ModulePath},
		{"io.github.rfizzle.go-starter.environment", annotationKey},
		{"Go Starter Project", cfg.ApplicationName},
		{defaultApplicationName, cfg.ApplicationName},
		{defaultEnvPrefix, cfg.EnvPrefix},
		{defaultBinaryName, cfg.BinaryName},
		{defaultProjectSlug, cfg.BinaryName},
	}

	switch filepath.ToSlash(path) {
	case ".goreleaser.yaml", "build/docker/Dockerfile",
		"build/docker/Dockerfile.release", "scripts/container-build.bash":
		replacements = append(replacements, replacement{defaultReleaseName, cfg.BinaryName})
	case ".dockerignore":
		replacements = append(replacements, replacement{"/" + defaultReleaseName, "/" + cfg.BinaryName})
	case "docs/development/deployment.md":
		replacements = append(replacements, replacement{
			"dist/" + defaultReleaseName + "-image.tar",
			"dist/" + cfg.BinaryName + "-image.tar",
		})
	}

	text = applyReplacements(text, replacements)

	description := strconv.Quote(cfg.Description)
	switch filepath.ToSlash(path) {
	case "api/openapi.yml":
		text = replaceLine(text, "  title: "+cfg.ApplicationName, "  title: "+strconv.Quote(cfg.ApplicationName)+"\n  description: "+description)
	case "build/docker/Dockerfile":
		text = replaceLine(text, `    org.opencontainers.image.description="OpenAPI-first foundation for PostgreSQL-backed Go services" \`, "    org.opencontainers.image.description="+description+` \`)
	case ".goreleaser.yaml":
		text = replaceLine(text, "      org.opencontainers.image.description: OpenAPI-first foundation for PostgreSQL-backed Go services", "      org.opencontainers.image.description: "+description)
	case "cmd/root.go":
		text = replaceLine(text, `	Short:   "A starter golang project",`, "\tShort:   "+description+",")
	}

	updated := []byte(text)
	if filepath.Ext(path) == ".go" && !bytes.Equal(contents, updated) {
		formatted, err := format.Source(updated)
		if err != nil {
			return nil, fmt.Errorf("formatting transformed Go source: %w", err)
		}
		updated = formatted
	}
	return updated, nil
}

// applyReplacements rewrites text in one left-to-right pass, so text produced by
// a replacement is never rescanned. At each position the longest matching token
// wins. Both properties are required because the template identifiers overlap
// (gostarter, go-starter, starter) and the substituted names may themselves
// contain a template token: sequential ReplaceAll passes would rewrite their own
// output, turning a --binary mystarter into mymystarter.
func applyReplacements(text string, replacements []replacement) string {
	ordered := slices.Clone(replacements)
	slices.SortStableFunc(ordered, func(a, b replacement) int {
		return len(b.old) - len(a.old)
	})

	var rewritten strings.Builder
	rewritten.Grow(len(text))
	for index := 0; index < len(text); {
		matched := false
		for _, candidate := range ordered {
			if candidate.old != "" && strings.HasPrefix(text[index:], candidate.old) {
				rewritten.WriteString(candidate.new)
				index += len(candidate.old)
				matched = true
				break
			}
		}
		if !matched {
			rewritten.WriteByte(text[index])
			index++
		}
	}
	return rewritten.String()
}

func replaceLine(text, oldLine, newLine string) string {
	return strings.Replace(text, oldLine, newLine, 1)
}

func sanitizeAnnotation(modulePath string) string {
	var result strings.Builder
	for _, r := range strings.ToLower(modulePath) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '.', r == '-':
			result.WriteRune(r)
		case r == '/', r == '_', r == '~':
			result.WriteByte('.')
		default:
			result.WriteByte('-')
		}
	}
	return strings.Trim(result.String(), ".-")
}

func rollback(changes []fileChange) error {
	var errs error
	for i := len(changes) - 1; i >= 0; i-- {
		change := changes[i]
		if err := writeFileAtomically(change.path, change.original, change.mode); err != nil {
			errs = errors.Join(errs, fmt.Errorf("rolling back %s: %w", change.path, err))
		}
	}
	return errs
}

func writeFileAtomically(path string, contents []byte, mode os.FileMode) (retErr error) {
	file, err := os.CreateTemp(filepath.Dir(path), ".template-init-*")
	if err != nil {
		return err
	}
	tempPath := file.Name()
	defer func() {
		if err := os.Remove(tempPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			retErr = errors.Join(retErr, fmt.Errorf("removing temporary file: %w", err))
		}
	}()

	if err := file.Chmod(mode); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(contents); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

func git(ctx context.Context, root string, args ...string) ([]byte, error) {
	commandArgs := append([]string{"-c", "core.fsmonitor=false", "-C", root}, args...)
	command := exec.CommandContext(ctx, "git", commandArgs...)
	out, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, bytes.TrimSpace(out))
	}
	return out, nil
}
