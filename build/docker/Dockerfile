# Multi-platform index digests are pinned for reproducible amd64 and arm64
# builds. Dependabot proposes weekly updates; see README.md for the validation
# steps required when either digest changes.
FROM --platform=$BUILDPLATFORM golang:1.26.5-alpine3.24@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS build

WORKDIR /src

ARG TARGETOS
ARG TARGETARCH
ARG BUILD_BRANCH
ARG BUILD_DATE
ARG BUILD_DATE_RFC3339
ARG BUILD_ENV
ARG BUILD_REVISION
ARG BUILD_VERSION

COPY go.mod go.sum ./
RUN go mod download
COPY main.go generate.go ./
COPY cmd ./cmd/
COPY internal ./internal/
COPY web ./web/

RUN test -n "${TARGETOS}" \
    && test -n "${TARGETARCH}" \
    && test -n "${BUILD_BRANCH}" \
    && test -n "${BUILD_DATE}" \
    && test -n "${BUILD_DATE_RFC3339}" \
    && test -n "${BUILD_ENV}" \
    && test -n "${BUILD_REVISION}" \
    && test -n "${BUILD_VERSION}" \
    && CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" \
    go build -trimpath \
    -ldflags="-s -w \
    -X github.com/rfizzle/go-starter/cmd.ApplicationName=starter \
    -X github.com/rfizzle/go-starter/cmd.BuildBranch=${BUILD_BRANCH} \
    -X github.com/rfizzle/go-starter/cmd.BuildDate=${BUILD_DATE} \
    -X github.com/rfizzle/go-starter/cmd.BuildEnv=${BUILD_ENV} \
    -X github.com/rfizzle/go-starter/cmd.BuildRevision=${BUILD_REVISION} \
    -X github.com/rfizzle/go-starter/cmd.BuildVersion=${BUILD_VERSION}" \
    -o /out/starter . \
    && mkdir -p /runtime/output /runtime/secrets /runtime/tmp /runtime/etc/app

FROM gcr.io/distroless/static-debian13:nonroot@sha256:f7f8f729987ad0fdf6b05eeeae94b26e6a0f613bdf46feea7fc40f7bd72953e6 AS final

ARG BUILD_BRANCH
ARG BUILD_DATE_RFC3339
ARG BUILD_ENV
ARG BUILD_REVISION
ARG BUILD_VERSION

LABEL org.opencontainers.image.created="${BUILD_DATE_RFC3339}" \
    org.opencontainers.image.description="OpenAPI-first foundation for PostgreSQL-backed Go services" \
    org.opencontainers.image.licenses="MIT" \
    org.opencontainers.image.ref.name="${BUILD_BRANCH}" \
    org.opencontainers.image.revision="${BUILD_REVISION}" \
    org.opencontainers.image.source="https://github.com/rfizzle/go-starter" \
    org.opencontainers.image.title="starter" \
    org.opencontainers.image.version="${BUILD_VERSION}" \
    io.github.rfizzle.go-starter.environment="${BUILD_ENV}"

COPY --chown=nonroot:nonroot --from=build /out/starter /app
COPY --chown=nonroot:nonroot --from=build /runtime/output /output
COPY --chown=nonroot:nonroot --from=build /runtime/secrets /secrets
COPY --chown=nonroot:nonroot --from=build /runtime/tmp /tmp
COPY --chown=nonroot:nonroot --from=build /runtime/etc/app /etc/app

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app"]
CMD ["serve"]
