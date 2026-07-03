package remote

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func (s *Service) loadStoredCredential(ctx context.Context) (types.Credential, error) {
	stored, err := s.d.Store.Org().GetIntegrationCredential(ctx)
	if err != nil {
		return types.Credential{}, err
	}
	if stored == nil {
		return types.Credential{}, domain.NewDomainError(domain.StatusUnprocessable, "data source not connected")
	}
	key, err := s.d.CredentialKey()
	if err != nil {
		return types.Credential{}, err
	}
	raw, err := common.Decrypt(key, stored.Encrypted)
	if err != nil {
		return types.Credential{}, domain.NewDomainError(domain.StatusUnprocessable, "failed to decrypt credential")
	}
	return types.UnmarshalCredentialPayload(stored.Platform, raw)
}

func (s *Service) saveCredential(ctx context.Context, cred types.Credential) error {
	key, err := s.d.CredentialKey()
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
	return s.d.Store.Org().SaveIntegrationCredential(ctx, cred.Platform, encrypted)
}

func (s *Service) providerForStored(ctx context.Context) (datasource.Provider, types.Platform, error) {
	cred, err := s.loadStoredCredential(ctx)
	if err != nil {
		return nil, "", err
	}
	provider, err := s.d.Factory.ForPlatform(cred.Platform, cred)
	if err != nil {
		return nil, "", err
	}
	return provider, cred.Platform, nil
}

func (s *Service) providerForCredential(cred types.Credential) (datasource.Provider, error) {
	return s.d.Factory.ForPlatform(cred.Platform, cred)
}
