package core

import (
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

const credentialKeySize = 32

func (d *Deps) CredentialKey() ([]byte, error) {
	if len(d.cryptoKey) == credentialKeySize {
		return d.cryptoKey, nil
	}
	key, err := common.ParseKey(d.Cfg.DataSourceCredentialKey)
	if err != nil {
		return nil, domain.NewDomainError(domain.StatusUnprocessable, "DATA_SOURCE_CREDENTIAL_KEY is required")
	}
	d.cryptoKey = key
	return key, nil
}
