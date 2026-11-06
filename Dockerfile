FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /sampatti ./cmd/sampatti

# Runtime
# alpine:3.20 is end-of-life (2026-04); its ca-certificates no longer get updates.
FROM alpine:3.24

# eMudhra intermediate CA key compromise (https://emudhra.we-have-a-problem.com/,
# disclosed 2026-10-20). ca-certificates 20260413 still trusts emSign ECC Root
# CA - G3; current packages don't. A stale cached layer of this RUN kept the old
# package, so build with --no-cache, and fail the build if the root is in the
# bundle or cert directory (Go reads both). The needle is a line of that root's
# PEM (SHA-256 86A1ECBA...BE6B).
RUN apk add --no-cache ca-certificates \
 && if grep -RqsF 'EwJJTjETMBEGA1UECxMKZW1TaWduIFBLSTElMCMGA1UEChMcZU11ZGhyYSBUZWNo' /etc/ssl/certs /etc/ssl/cert.pem; then \
      echo "emSign ECC Root CA - G3 is trusted; rebuild with --no-cache" >&2; exit 1; \
    fi

COPY --from=builder /sampatti /usr/local/bin/sampatti

EXPOSE 8081

ENTRYPOINT ["sampatti"]
