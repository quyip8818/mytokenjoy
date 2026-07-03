package config

import "strings"

const (
	DefaultCompanySlug     = "default"
	SaaSDefaultCompanyName = "Test Company"
)

func (c Config) ResolvedCompanyName() string {
	if c.SupportSaas {
		return SaaSDefaultCompanyName
	}
	return strings.TrimSpace(c.CompanyName)
}
