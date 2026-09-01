# syntax=docker/dockerfile:1

# ---------------------------------------------------------------------------
# Builder stage: compiles the statically-linked queryx CLI binary.
# ---------------------------------------------------------------------------
ARG GO_VERSION=1.27
FROM golang:${GO_VERSION}-alpine AS builder

# git is required by the Go toolchain for module operations.
RUN apk add --no-cache git

WORKDIR /src

# Download dependencies first so this layer is cached across source changes.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source and build a static binary.
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/queryx ./cmd/queryx

# ---------------------------------------------------------------------------
# Runtime stage: tiny image containing only the binary and CA certificates
# (needed for HTTPS queries, e.g. CFX.re / FiveM endpoints behind TLS).
# ---------------------------------------------------------------------------
FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates && adduser -D -u 10001 queryx

COPY --from=builder /out/queryx /usr/local/bin/queryx

USER queryx
ENTRYPOINT ["queryx"]
CMD ["-version"]
