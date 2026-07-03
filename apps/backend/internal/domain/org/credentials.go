package org

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

const credentialKeySize = 32

func (s *service) credentialKey() ([]byte, error) {
	if len(s.cryptoKey) == credentialKeySize {
		return s.cryptoKey, nil
	}
	if key, err := common.ParseKey(s.cfg.DataSourceCredentialKey); err == nil {
		s.cryptoKey = key
		return key, nil
	}
	if s.cfg.IsDemoProfile() {
		s.cryptoKey = common.DevDefaultKey()
		return s.cryptoKey, nil
	}
	return nil, domain.NewDomainError(domain.StatusUnprocessable, "DATA_SOURCE_CREDENTIAL_KEY is required")
}

func (s *service) loadStoredCredential(ctx context.Context) (types.Credential, error) {
	stored, err := s.store.Org().GetIntegrationCredential(ctx)
	if err != nil {
		return types.Credential{}, err
	}
	if stored == nil {
		return types.Credential{}, domain.NewDomainError(domain.StatusUnprocessable, "data source not connected")
	}
	key, err := s.credentialKey()
	if err != nil {
		return types.Credential{}, err
	}
	raw, err := common.Decrypt(key, stored.Encrypted)
	if err != nil {
		return types.Credential{}, domain.NewDomainError(domain.StatusUnprocessable, "failed to decrypt credential")
	}
	return types.UnmarshalCredentialPayload(stored.Platform, raw)
}

func (s *service) saveCredential(ctx context.Context, cred types.Credential) error {
	key, err := s.credentialKey()
	if err != nil {
		return err
	}
	payload, err := types.MarshalCredentialPayload(cred)
	if err != nil {
		return err
	}
	encrypted, err := common.Encrypt(key, payload)
	if err != nil {
		return err
	}
	return s.store.Org().SaveIntegrationCredential(ctx, cred.Platform, encrypted)
}

func (s *service) providerForStored(ctx context.Context) (datasource.Provider, types.Platform, error) {
	cred, err := s.loadStoredCredential(ctx)
	if err != nil {
		return nil, "", err
	}
	provider, err := s.factory.ForPlatform(cred.Platform, cred)
	if err != nil {
		return nil, "", err
	}
	return provider, cred.Platform, nil
}

func (s *service) providerForCredential(cred types.Credential) (datasource.Provider, error) {
	return s.factory.ForPlatform(cred.Platform, cred)
}
