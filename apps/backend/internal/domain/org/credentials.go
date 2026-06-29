package org

import (
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/internal/pkg/cryptoutil"
)

const credentialKeySize = 32

func (s *service) credentialKey() ([]byte, error) {
	if len(s.cryptoKey) == credentialKeySize {
		return s.cryptoKey, nil
	}
	if key, err := cryptoutil.ParseKey(s.cfg.DataSourceCredentialKey); err == nil {
		s.cryptoKey = key
		return key, nil
	}
	if s.cfg.IsDemoProfile() {
		s.cryptoKey = cryptoutil.DevDefaultKey()
		return s.cryptoKey, nil
	}
	return nil, domain.NewDomainError(domain.StatusUnprocessable, "DATA_SOURCE_CREDENTIAL_KEY is required")
}

func (s *service) loadStoredCredential() (types.Credential, error) {
	stored, err := s.store.Credential().GetCredential()
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
	raw, err := cryptoutil.Decrypt(key, stored.Encrypted)
	if err != nil {
		return types.Credential{}, domain.NewDomainError(domain.StatusUnprocessable, "failed to decrypt credential")
	}
	return types.UnmarshalCredentialPayload(stored.Platform, raw)
}

func (s *service) saveCredential(cred types.Credential) error {
	key, err := s.credentialKey()
	if err != nil {
		return err
	}
	payload, err := types.MarshalCredentialPayload(cred)
	if err != nil {
		return err
	}
	encrypted, err := cryptoutil.Encrypt(key, payload)
	if err != nil {
		return err
	}
	return s.store.Credential().SaveCredential(cred.Platform, encrypted)
}

func (s *service) providerForStored() (datasource.Provider, types.Platform, error) {
	cred, err := s.loadStoredCredential()
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
