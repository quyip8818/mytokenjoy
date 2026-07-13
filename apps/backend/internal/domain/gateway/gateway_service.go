package gateway

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/tokenjoy/backend/internal/config"
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
}

func NewGatewayService(cfg config.Config, precheck Prechecker) (GatewayService, error) {
	target, err := url.Parse(strings.TrimRight(cfg.NewAPIBaseURL, "/"))
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &http.Transport{DisableCompression: true}
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}
	return &gatewayService{
		precheck:      precheck,
		proxy:         proxy,
		allowDevModel: cfg.AllowsDevHTTPRoutes(),
	}, nil
}

func (g *gatewayService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !isAllowedGatewayPath(r.URL.Path) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer sk-") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	platformKeySecret := strings.TrimPrefix(auth, "Bearer ")
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
	if !g.allowDevModel && isDevOnlyModel(model) {
		http.Error(w, "request rejected", http.StatusForbidden)
		return
	}
	if err := g.precheck.Run(
		r.Context(),
		store.HashPlatformKey(platformKeySecret),
		model,
		r.URL.Path == "/v1/models",
	); err != nil {
		http.Error(w, "request rejected", http.StatusForbidden)
		return
	}
	g.proxy.ServeHTTP(w, r)
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

// DevOnlyModel is the catalog model backed by the local dev-mock upstream
// (see apps/dev-mock-llm). It is only reachable when DEPLOY_ENV=local.
const DevOnlyModel = "local-test-model"

func isDevOnlyModel(model string) bool {
	return model == DevOnlyModel
}

var _ GatewayService = (*gatewayService)(nil)
