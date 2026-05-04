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
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=builder /sampatti /usr/local/bin/sampatti

EXPOSE 8081

ENTRYPOINT ["sampatti"]
