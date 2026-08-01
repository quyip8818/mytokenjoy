# apps-backend: Go multi-stage build
# Context: repo root
FROM golang:1.25-alpine AS builder
RUN sed -i 's#https\?://dl-cdn.alpinelinux.org#https://mirrors.aliyun.com#g' /etc/apk/repositories \
    && apk add --no-cache git ca-certificates
ENV GOPROXY=https://goproxy.cn,direct
ENV GOMEMLIMIT=1GiB
WORKDIR /build
COPY apps/backend/go.mod apps/backend/go.sum ./
RUN go mod download
COPY apps/backend/ .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -p 1 -o server ./cmd/server

FROM alpine:3.21
RUN sed -i 's#https\?://dl-cdn.alpinelinux.org#https://mirrors.aliyun.com#g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S app -G app
COPY --from=builder /build/server /usr/local/bin/server
USER app
EXPOSE 8010
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s \
  CMD wget -qO- http://localhost:8010/healthz || exit 1
ENTRYPOINT ["server"]
