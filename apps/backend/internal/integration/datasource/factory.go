package datasource

import (
	"net/http"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource/feishu"
)

type factory struct {
	cfg        config.Config
	httpClient *http.Client
}

func NewFactory(cfg config.Config) Factory {
	return &factory{
		cfg:        cfg,
		httpClient: &http.Client{},
	}
}

func (f *factory) ForPlatform(platform types.Platform, cred types.Credential) (Provider, error) {
	switch platform {
	case types.PlatformFeishu:
		if cred.Feishu == nil {
			return nil, domain.NewDomainError(domain.StatusUnprocessable, "invalid feishu credential")
		}
		return newFeishuProvider(feishu.NewClient(f.cfg.FeishuBaseURL, *cred.Feishu, f.httpClient)), nil
	case types.PlatformDingtalk, types.PlatformWecom:
		return nil, domain.NewDomainError(domain.StatusUnprocessable, "platform not supported")
	default:
		return nil, domain.NewDomainError(domain.StatusUnprocessable, "unsupported platform")
	}
}

var _ Factory = (*factory)(nil)
