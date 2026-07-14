ARG UPSTREAM_REF=1e80ce03e8592af09eafb32c36ac20e79d359937

FROM oven/bun:1-alpine@sha256:5acc90a93e91ff07bf72aa90a7c9f0fa189765aec90b47bdbf2152d2196383c0 AS upstream-src
ARG UPSTREAM_REF
RUN apk add --no-cache git ca-certificates
WORKDIR /upstream
RUN git clone https://github.com/QuantumNous/new-api.git . \
    && git fetch --depth 1 origin "${UPSTREAM_REF}" \
    && git checkout "${UPSTREAM_REF}"

FROM upstream-src AS web-default
WORKDIR /upstream/web
RUN bun install --frozen-lockfile \
    && cd default \
    && DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(cat /upstream/VERSION) bun run build

FROM upstream-src AS web-classic
WORKDIR /upstream/web
RUN bun install --frozen-lockfile \
    && cd classic \
    && VITE_REACT_APP_VERSION=$(cat /upstream/VERSION) bun run build

FROM golang:1.26.1-alpine@sha256:2389ebfa5b7f43eeafbd6be0c3700cc46690ef842ad962f6c5bd6be49ed82039 AS go-builder
ARG UPSTREAM_REF
ENV GO111MODULE=on CGO_ENABLED=0 GOEXPERIMENT=greenteagc

ARG TARGETOS
ARG TARGETARCH
ENV GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64}

RUN apk add --no-cache git ca-certificates

WORKDIR /build
RUN git clone https://github.com/QuantumNous/new-api.git . \
    && git fetch --depth 1 origin "${UPSTREAM_REF}" \
    && git checkout "${UPSTREAM_REF}"

COPY patches/management/notify.go patches/management/notify_test.go ./management/
COPY patches/new-api/0001-management-webhook.patch /tmp/0001-management-webhook.patch
RUN git apply /tmp/0001-management-webhook.patch \
    && go test -count=1 ./management/...

COPY patches/new-api/0002-admin-token-contract.patch /tmp/0002-admin-token-contract.patch
RUN git apply /tmp/0002-admin-token-contract.patch

COPY --from=web-default /upstream/web/default/dist ./web/default/dist
COPY --from=web-classic /upstream/web/classic/dist ./web/classic/dist

RUN go build -ldflags "-s -w -X 'github.com/QuantumNous/new-api/common.Version=$(cat VERSION)'" -o new-api

FROM debian:bookworm-slim@sha256:f06537653ac770703bc45b4b113475bd402f451e85223f0f2837acbf89ab020a

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata libasan8 wget curl \
    && rm -rf /var/lib/apt/lists/* \
    && update-ca-certificates

COPY --from=go-builder /build/new-api /new-api
COPY --from=go-builder /build/LICENSE /build/NOTICE /build/THIRD-PARTY-LICENSES.md /licenses/

EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/new-api"]
