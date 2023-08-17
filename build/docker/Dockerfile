# syntax=docker/dockerfile:1

# Accept the Go version for the image to be set as a build argument.
# Default to Go 1.19
ARG GO_VERSION=1.19

# Build stage: Build the binary
FROM golang:${GO_VERSION}-alpine as golang

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git make zip bash

# Make empty directory for permission copy
RUN mkdir /empty

# Set the working directory outside $GOPATH to enable the support for modules.
WORKDIR /src

# Copy Go Module config
COPY go.mod .
COPY go.sum .
COPY vendor ./vendor

# Download Go Modules
RUN go mod download -x
RUN go mod verify

# Import the code from the context.
COPY . .

# Build app
RUN make linux

# Set permissions on app
RUN chmod +x /src/bin/linux/starter

# Final stage: Bare minimum run container.
FROM gcr.io/distroless/static as final

# Copy binary
COPY --chown=nonroot:nonroot --from=golang /src/bin/linux/starter /app

# Create some useful empty directories
COPY --chown=nonroot:nonroot --from=golang /empty /output
COPY --chown=nonroot:nonroot --from=golang /empty /secrets

# Copy temp directory
COPY --chown=nonroot:nonroot --from=golang /empty /tmp

# Copy etc directory
COPY --chown=nonroot:nonroot --from=golang /empty /etc

# Copy collector config directory
COPY --chown=nonroot:nonroot --from=golang /empty /etc/app

# Perform any further action as an unprivileged user.
USER nonroot:nonroot

# Set port exposure
EXPOSE 8080

# Set the entrypoint of the container
CMD ["/app", "serve"]
