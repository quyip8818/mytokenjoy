package config

import "strings"

const (
	SaaSDefaultCompanyName = "Test Company"
)

func (c Config) ResolvedCompanyName() string {
	if c.SupportSaas {
		return SaaSDefaultCompanyName
	}
	return strings.TrimSpace(c.CompanyName)
}
