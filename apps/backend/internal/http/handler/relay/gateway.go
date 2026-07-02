package relay

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/tokenjoy/backend/internal/config"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/store"
)

type Gateway struct {
	cfg         config.Config
	store       store.Store
	precheck    relay.Prechecker
	proxyTarget *url.URL
}

func NewGateway(cfg config.Config, st store.Store, precheck relay.Prechecker) (*Gateway, error) {
	target, err := url.Parse(strings.TrimRight(cfg.NewAPIBaseURL, "/"))
	if err != nil {
		return nil, err
	}
	return &Gateway{cfg: cfg, store: st, precheck: precheck, proxyTarget: target}, nil
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer sk-") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tokenKey := strings.TrimPrefix(auth, "Bearer ")
	mapping, err := g.store.Relay().GetMappingByFullKey(r.Context(), tokenKey)
	if err != nil || mapping == nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	ctx := domaincompany.WithContext(r.Context(), domaincompany.Context{CompanyID: mapping.CompanyID})
	company, err := g.store.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil || company == nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if company.NewAPIWalletUserID != nil {
		ctx = domaincompany.WithContext(ctx, domaincompany.Context{
			CompanyID:          mapping.CompanyID,
			NewAPIWalletUserID: *company.NewAPIWalletUserID,
			Status:             company.Status,
		})
	}
	body, err := readAndRestoreBody(r)
	if err != nil {
		http.Error(w, "read request body", http.StatusForbidden)
		return
	}
	if err := g.precheck.Run(ctx, relay.PrecheckInput{
		Mapping: mapping,
		Company: company,
		Model:   parseRequestModel(body),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	g.proxy(w, r)
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

func (g *Gateway) proxy(w http.ResponseWriter, r *http.Request) {
	targetURL := *g.proxyTarget
	targetURL.Path = strings.TrimPrefix(r.URL.Path, "/v1")
	if targetURL.Path == "" {
		targetURL.Path = "/"
	}
	targetURL.RawQuery = r.URL.RawQuery
	proxy := httputil.NewSingleHostReverseProxy(&targetURL)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("Authorization", r.Header.Get("Authorization"))
	}
	proxy.Transport = &http.Transport{DisableCompression: true}
	proxy.ServeHTTP(w, r)
}
