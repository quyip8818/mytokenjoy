package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/tokenjoy/backend/internal/config"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
)

type RelayMappingReader interface {
	GetMappingByFullKey(ctx context.Context, tokenKey string) (*store.RelayMapping, error)
}

type CompanyReader interface {
	GetByID(ctx context.Context, companyID int64) (*store.Company, error)
}

type GatewayService interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type gatewayService struct {
	cfg         config.Config
	mappings    RelayMappingReader
	companies   CompanyReader
	precheck    Prechecker
	proxyTarget *url.URL
}

func NewGatewayService(
	cfg config.Config,
	mappings RelayMappingReader,
	companies CompanyReader,
	precheck Prechecker,
) (GatewayService, error) {
	target, err := url.Parse(strings.TrimRight(cfg.NewAPIBaseURL, "/"))
	if err != nil {
		return nil, err
	}
	return &gatewayService{
		cfg:         cfg,
		mappings:    mappings,
		companies:   companies,
		precheck:    precheck,
		proxyTarget: target,
	}, nil
}

func (g *gatewayService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer sk-") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tokenKey := strings.TrimPrefix(auth, "Bearer ")
	mapping, err := g.mappings.GetMappingByFullKey(r.Context(), tokenKey)
	if err != nil || mapping == nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	ctx := domaincompany.WithContext(r.Context(), domaincompany.Context{CompanyID: mapping.CompanyID})
	company, err := g.companies.GetByID(ctx, mapping.CompanyID)
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
	if err := g.precheck.Run(ctx, PrecheckInput{
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

func (g *gatewayService) proxy(w http.ResponseWriter, r *http.Request) {
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

var _ GatewayService = (*gatewayService)(nil)
