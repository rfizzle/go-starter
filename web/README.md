# Embedded frontend boundary

Go Starter is backend-first. It does not select, install, build, or test a
JavaScript framework. The committed `build/index.html` is a minimal placeholder
that keeps `go build` and container builds independent of Node and package
manager tooling.

The Go package embeds every committed file matched by `build/*`. The HTTP server
serves those files only outside the reserved `/api`, `/metrics`, and `/debug`
namespaces, so a frontend fallback cannot shadow API, Prometheus, or pprof
routes. A missing static path returns `404`; the starter does not provide SPA
history fallback.

## Adding a frontend

A downstream Vite, Svelte, React, or other frontend can live under `web/` and
write its production output to `web/build/`. For Vite, set `build.outDir` to
`build` relative to `web/vite.config.*`. Keep frontend source and package
metadata under `web/` rather than turning the repository root into a JavaScript
package solely for command aliases.

Choose an asset policy explicitly:

- Commit production assets in `web/build/` so Go-only builds remain
  self-contained.
- Or make the frontend build an explicit prerequisite in local and CI build
  workflows, while retaining a committed placeholder file for clean-checkout
  compilation.

If the frontend uses client-side routing, add and test an intentional SPA
history fallback without changing the reserved backend route namespaces.
