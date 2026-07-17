package gateway

import (
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
	"github.com/tokenjoy/backend/internal/infra/ratelimit"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
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
}

func NewGatewayService(cfg config.Config, precheck Prechecker, limiter ratelimit.Limiter, logger *slog.Logger) (GatewayService, error) {
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
	return &gatewayService{
		precheck:      precheck,
		proxy:         proxy,
		allowDevModel: cfg.AllowsDevHTTPRoutes(),
		rateLimiter:   limiter,
		rlRate:        cfg.RateLimitV1Rate,
		rlBurst:       cfg.RateLimitV1Burst,
		rlDryRun:      cfg.RateLimitDryRun,
		logger:        logger,
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
	body, err := readAndRestoreBody(r)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "read request body", http.StatusForbidden)
		return
	}
	model := parseRequestModel(body)
	if !g.allowDevModel && modelcatalog.IsLocalOnlyCallType(model) {
		logGatewayRejection(r.URL.Path, model, "dev-only model outside local environment")
		http.Error(w, "request rejected", http.StatusForbidden)
		return
	}
	opts := PrecheckForRequest(r.URL.Path, model, g.allowDevModel)
	_, err = g.precheck.Run(r.Context(), keyHash, model, opts)
	if err != nil {
		logGatewayRejection(r.URL.Path, model, err.Error())
		http.Error(w, "request rejected", http.StatusForbidden)
		return
	}
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

func readAndRestoreBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	r.ContentLength = int64(len(body))
	return body, nil
}

func parseRequestModel(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var payload struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	return payload.Model
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
