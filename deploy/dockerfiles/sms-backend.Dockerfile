# sms-backend: Go HTTP server
# Build context: repo root (docker build -f deploy/dockerfiles/sms-backend.Dockerfile .)
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /build
COPY sms/backend/go.mod sms/backend/go.sum ./
RUN go mod download
COPY sms/backend/ .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S app -G app
COPY --from=builder /build/server /usr/local/bin/server
USER app
EXPOSE 8020
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
  CMD wget -qO- http://localhost:8020/health || exit 1
ENTRYPOINT ["server"]
