package store

import (
	"github.com/google/uuid"
)

const (
	RebalanceAxisMember  = "member"
	RebalanceAxisProject = "project"
	RebalanceAxisCompany = "company"
)

// CompanyAxisID formats a company ID as the axis identifier used by rebalance jobs.
func CompanyAxisID(companyID uuid.UUID) string {
	return companyID.String()
}
