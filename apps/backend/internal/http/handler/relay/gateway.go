package relayhandler

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
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

const gatewayMinEstimateCNY = 0.01

type Gateway struct {
	cfg         config.Config
	store       store.Store
	wallet      domaincompany.WalletService
	proxyTarget *url.URL
}

func NewGateway(cfg config.Config, st store.Store, wallet domaincompany.WalletService) (*Gateway, error) {
	target, err := url.Parse(strings.TrimRight(cfg.NewAPIBaseURL, "/"))
	if err != nil {
		return nil, err
	}
	return &Gateway{cfg: cfg, store: st, wallet: wallet, proxyTarget: target}, nil
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
	if company.NewAPIWalletAccountID != nil {
		ctx = domaincompany.WithContext(ctx, domaincompany.Context{
			CompanyID:             mapping.CompanyID,
			NewAPIWalletAccountID: *company.NewAPIWalletAccountID,
			Status:                company.Status,
		})
	}
	if err := g.precheck(ctx, mapping, company, r); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	g.proxy(w, r)
}

func (g *Gateway) precheck(ctx context.Context, mapping *store.RelayMapping, company *store.Company, r *http.Request) error {
	if company.Status != store.CompanyStatusActive {
		return fmt.Errorf("company not active")
	}
	if err := g.checkWallet(ctx, company); err != nil {
		return err
	}
	if err := g.checkDepartmentBudget(ctx, mapping); err != nil {
		return err
	}
	if err := g.checkTokenRemainQuota(mapping); err != nil {
		return err
	}
	body, err := readAndRestoreBody(r)
	if err != nil {
		return fmt.Errorf("read request body")
	}
	return g.checkPlatformKey(ctx, mapping, parseRequestModel(body))
}

func (g *Gateway) checkWallet(ctx context.Context, company *store.Company) error {
	if company.NewAPIWalletAccountID == nil || g.wallet == nil {
		return nil
	}
	quota, err := g.wallet.AvailableQuota(ctx, *company.NewAPIWalletAccountID)
	if err != nil {
		return fmt.Errorf("wallet unavailable")
	}
	balanceCNY := newapi.FromNewAPIUnits(quota, nil, nil)
	if balanceCNY < gatewayMinEstimateCNY {
		return fmt.Errorf("insufficient wallet balance")
	}
	return nil
}

func (g *Gateway) checkDepartmentBudget(ctx context.Context, mapping *store.RelayMapping) error {
	tree, err := g.store.Budget().Tree(ctx)
	if err != nil {
		return err
	}
	node := pkgbudget.FindBudgetNode(tree, mapping.DepartmentID)
	if node == nil {
		return fmt.Errorf("department not found")
	}
	if node.Budget <= 0 {
		return fmt.Errorf("budget exceeded")
	}
	if node.Consumed+gatewayMinEstimateCNY > node.Budget {
		return fmt.Errorf("budget exceeded")
	}
	return nil
}

func (g *Gateway) checkTokenRemainQuota(mapping *store.RelayMapping) error {
	if mapping.RelayRemainQuota == nil || *mapping.RelayRemainQuota <= 0 {
		return fmt.Errorf("insufficient token quota")
	}
	return nil
}

func (g *Gateway) checkPlatformKey(ctx context.Context, mapping *store.RelayMapping, modelName string) error {
	keys, err := g.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return err
	}
	var key *types.PlatformKey
	for i := range keys {
		if keys[i].ID == mapping.PlatformKeyID {
			key = &keys[i]
			break
		}
	}
	if key == nil {
		return fmt.Errorf("platform key not found")
	}
	if key.Status != "active" {
		return fmt.Errorf("platform key inactive")
	}
	if modelName == "" || len(key.ModelWhitelist) == 0 {
		return nil
	}
	for _, allowed := range key.ModelWhitelist {
		if allowed == modelName {
			return nil
		}
	}
	return fmt.Errorf("model not allowed")
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
