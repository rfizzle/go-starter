# Go Starter

[![CI](https://github.com/rfizzle/go-starter/actions/workflows/ci.yml/badge.svg)](https://github.com/rfizzle/go-starter/actions/workflows/ci.yml)
[![Go 1.26.5](https://img.shields.io/badge/Go-1.26.5-00ADD8?logo=go&logoColor=white)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue)](LICENSE)

Go Starter is a template for OpenAPI-first, PostgreSQL-backed Go HTTP services:
clone it, give it your names once, and build your service on it.

Status: pre-release; no tagged version yet.

## What you get

- Cobra and Viper commands (`serve`, `migrate`, `seed`) with layered configuration validated before anything opens
- Chi routing and ogen types generated from `api/openapi.yml`
- PostgreSQL through pgx/v5, sqlc and Goose, with serialized embedded migrations
- JWT bearer authentication with an example user module and login endpoint
- Health endpoints, Prometheus metrics, Zap logging and optional OpenTelemetry export
- Login rate limiting, per process; forwarding headers are trusted only from `rate-limit.trusted-proxies`
- Unit tests and Testcontainers PostgreSQL integration tests, a CI workflow, GoReleaser and a Podman image build

It is a starter, not a deployment to ship unchanged: choose your secrets, CORS
origins, connection limits, TLS termination and monitoring first.

## Initialize a new project

Run the initializer once in a clean clone. It rewrites the module path, names,
environment prefix, image and OpenAPI metadata, then deletes itself:

```sh
make init \
  module=example.com/acme/orders \
  app-name='Acme Orders' \
  binary=orders-api \
  env-prefix=ORDERS \
  image=registry.example.com/acme/orders-api \
  description='Order processing API'
```

The description is optional. Review the diff, then run `make check`.

## Five-minute quickstart

You need Git, Go 1.26.5, Make, and Podman with a Compose provider.

```sh
git clone https://github.com/rfizzle/go-starter.git
cd go-starter
make tools
make env
cp .env.example .env
make dev-run
```

`dev-run` starts PostgreSQL, applies the migrations, seeds the development user
from `.env`, and serves on port 8080. In another terminal:

```sh
curl --fail http://localhost:8080/api/v1/healthz/readiness
curl --fail-with-body \
  --header 'Content-Type: application/json' \
  --data '{"email":"admin@example.com","password":"<seed-password>"}' \
  http://localhost:8080/api/v1/login
```

Stop it with Ctrl-C and `make dev-down`. Settings come from flags, then
`GOSTARTER_*` environment variables, then the file given with `--config`;
`./bin/gostarter serve --help` (after `make build`) lists them. The credentials
in `compose.yaml` and `.env.example` are for development only.

## Documents

[AGENTS.md](AGENTS.md) has the commands, the layout and the rules for anyone
changing the code; the design and the guides are in `docs/`, starting at
[docs/README.md](docs/README.md). Deployment is in
[docs/development/deployment.md](docs/development/deployment.md).
[CONTRIBUTING.md](CONTRIBUTING.md), [SECURITY.md](SECURITY.md) and
[CHANGELOG.md](CHANGELOG.md) cover contributing, vulnerability reports and
notable changes.

Go Starter is available under the [MIT License](LICENSE).
