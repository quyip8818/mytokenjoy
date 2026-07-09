package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
)

type RelayMappingReader interface {
	GetMappingByKeyHash(ctx context.Context, keyHash string) (*store.RelayMapping, error)
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
	if !isAllowedGatewayPath(r.URL.Path) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer sk-") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tokenKey := strings.TrimPrefix(auth, "Bearer ")
	mapping, err := g.mappings.GetMappingByKeyHash(r.Context(), store.HashPlatformKey(tokenKey))
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
		if domainErr, ok := err.(*domain.DomainError); ok {
			if domainErr.RetryAfter != nil {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", *domainErr.RetryAfter))
			}
			status := domainErr.Status
			if status == 0 {
				status = http.StatusForbidden
			}
			http.Error(w, domainErr.Message, status)
			return
		}
		http.Error(w, "request rejected", http.StatusForbidden)
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

var allowedGatewayPrefixes = []string{
	"/v1/chat/completions",
	"/v1/completions",
	"/v1/embeddings",
	"/v1/models",
}

func isAllowedGatewayPath(path string) bool {
	for _, prefix := range allowedGatewayPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func (g *gatewayService) proxy(w http.ResponseWriter, r *http.Request) {
	targetURL := *g.proxyTarget
	proxy := httputil.NewSingleHostReverseProxy(&targetURL)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = g.proxyTarget.Scheme
		req.URL.Host = g.proxyTarget.Host
		req.URL.Path = r.URL.Path
		req.URL.RawQuery = r.URL.RawQuery
		req.Host = g.proxyTarget.Host
		req.Header.Set("Authorization", r.Header.Get("Authorization"))
	}
	proxy.Transport = &http.Transport{DisableCompression: true}
	proxy.ServeHTTP(w, r)
}

var _ GatewayService = (*gatewayService)(nil)
