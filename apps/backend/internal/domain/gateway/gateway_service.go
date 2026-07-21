package gateway

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/infra/gatewaymetrics"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
	"github.com/tokenjoy/backend/internal/pkg/ratelimit"
	"github.com/tokenjoy/backend/internal/store"
)

const gatewayMaxBodyBytes = 4 << 20

type GatewayService interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type gatewayService struct {
	precheck      Prechecker
	proxy         *httputil.ReverseProxy
	allowDevModel bool
	rateLimiter   ratelimit.Limiter
	rlRate        int
	rlBurst       int
	rlDryRun      bool
	logger        *slog.Logger
	metrics       gatewaymetrics.Recorder
}

func NewGatewayService(cfg config.Config, precheck Prechecker, limiter ratelimit.Limiter, logger *slog.Logger, metrics gatewaymetrics.Recorder) (GatewayService, error) {
	target, err := url.Parse(cfg.NewAPIBaseURL)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
		TLSHandshakeTimeout:   5 * time.Second,
		DisableCompression:    true,
		MaxIdleConns:          200,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		ResponseHeaderTimeout: 120 * time.Second,
		ForceAttemptHTTP2:     true,
	}
	proxy.FlushInterval = -1 // Stream responses (SSE) in real-time.
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}
	if metrics == nil {
		metrics = gatewaymetrics.Noop()
	}
	return &gatewayService{
		precheck:      precheck,
		proxy:         proxy,
		allowDevModel: cfg.AllowsDevHTTPRoutes(),
		rateLimiter:   limiter,
		rlRate:        cfg.RateLimitV1Rate,
		rlBurst:       cfg.RateLimitV1Burst,
		rlDryRun:      cfg.RateLimitDryRun,
		logger:        logger,
		metrics:       metrics,
	}, nil
}

func (g *gatewayService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !isAllowedGatewayPath(r.URL.Path) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	platformKeySecret, ok := parseBearerSecret(r.Header.Get("Authorization"))
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Rate limit per API key (before precheck to save DB call on rejected requests).
	keyHash := store.HashPlatformKey(platformKeySecret)
	if !g.checkRateLimit(keyHash, w, r) {
		return
	}
	if r.Body != nil {
		r.Body = http.MaxBytesReader(w, r.Body, gatewayMaxBodyBytes)
	}
	model, err := peekModelFromBody(r)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "read request body", http.StatusForbidden)
		return
	}
	if !g.allowDevModel && modelcatalog.IsTestOnlyCallType(model) {
		g.metrics.RecordRejected()
		logGatewayRejection(r.URL.Path, model, "test-only model outside local environment")
		http.Error(w, "request rejected", http.StatusForbidden)
		return
	}
	opts := PrecheckForRequest(r.URL.Path, model, g.allowDevModel)
	start := time.Now()
	_, err = g.precheck.Run(r.Context(), keyHash, model, opts)
	g.metrics.RecordPrecheckDuration(time.Since(start))
	if err != nil {
		g.metrics.RecordRejected()
		logGatewayRejection(r.URL.Path, model, err.Error())
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	g.metrics.RecordAllowed()
	g.proxy.ServeHTTP(w, r)
}

// checkRateLimit applies per-key rate limiting. Returns true if the request is allowed.
func (g *gatewayService) checkRateLimit(keyHash string, w http.ResponseWriter, r *http.Request) bool {
	if g.rateLimiter == nil || !g.rlEnabled() {
		return true
	}
	key := fmt.Sprintf("rl:v1:%s", keyHash)
	result, err := g.rateLimiter.AllowTokenBucket(r.Context(), key, g.rlRate, g.rlBurst)
	if err != nil {
		// Fail-open on Redis error.
		if g.logger != nil {
			g.logger.Warn("rate_limit: v1 redis error, fail-open", "error", err, "key_prefix", keyHash[:8])
		}
		return true
	}
	ratelimit.WriteHeaders(w, result)
	if !result.Allowed {
		if g.rlDryRun {
			if g.logger != nil {
				g.logger.Warn("rate_limit: v1 would reject (dry-run)", "key_prefix", keyHash[:8])
			}
			return true
		}
		g.metrics.RecordRateLimited()
		ratelimit.WriteRejection(w, result)
		return false
	}
	return true
}

func (g *gatewayService) rlEnabled() bool {
	return g.rlRate > 0 && g.rlBurst > 0
}

func logGatewayRejection(path, model, reason string) {
	slog.Default().Info("gateway request rejected", "path", path, "model", model, "reason", reason)
}

// peekModelFromBody extracts the "model" field from the request body without
// consuming the entire body in the common case. It peeks up to peekSize bytes;
// if model is found there, the proxy gets the body via the buffered reader.
// If not found (e.g. large embedding input before model field), falls back to
// full body read — same cost as the old implementation but only on the edge case.
const peekSize = 4096

func peekModelFromBody(r *http.Request) (string, error) {
	if r.Body == nil {
		return "", nil
	}
	br := bufio.NewReaderSize(r.Body, peekSize)
	peeked, err := br.Peek(peekSize)
	if err != nil && err != io.EOF && err != bufio.ErrBufferFull {
		return "", err
	}

	if model := parseModelFromPrefix(peeked); model != "" {
		// Found in prefix — reassemble body without full read.
		r.Body = io.NopCloser(br)
		return model, nil
	}

	// Fallback: model not in first 4KB (rare — e.g. large embedding input before model).
	// Read the rest and try the full body.
	rest, err := io.ReadAll(br)
	if err != nil {
		return "", err
	}
	r.Body = io.NopCloser(bytes.NewReader(rest))
	r.ContentLength = int64(len(rest))
	return parseModelFromPrefix(rest), nil
}

// parseModelFromPrefix extracts "model" value from a JSON prefix (may be incomplete JSON).
func parseModelFromPrefix(data []byte) string {
	// Fast path: use json.Decoder token-by-token scan.
	dec := json.NewDecoder(bytes.NewReader(data))
	// Find opening {
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return ""
	}
	for dec.More() {
		// Read key
		keyTok, err := dec.Token()
		if err != nil {
			return ""
		}
		key, ok := keyTok.(string)
		if !ok {
			return ""
		}
		if key == "model" {
			valTok, err := dec.Token()
			if err != nil {
				return ""
			}
			if s, ok := valTok.(string); ok {
				return s
			}
			return ""
		}
		// Skip value — may fail on truncated JSON, that's fine.
		var discard json.RawMessage
		if err := dec.Decode(&discard); err != nil {
			return ""
		}
	}
	return ""
}

var allowedGatewayPaths = map[string]struct{}{
	"/v1/chat/completions": {},
	"/v1/completions":      {},
	"/v1/embeddings":       {},
	"/v1/models":           {},
}

func isAllowedGatewayPath(path string) bool {
	_, ok := allowedGatewayPaths[path]
	return ok
}

var _ GatewayService = (*gatewayService)(nil)
