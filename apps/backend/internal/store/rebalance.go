package store

import "strconv"

const (
	RebalanceAxisMember  = "member"
	RebalanceAxisProject = "project"
	RebalanceAxisCompany = "company"
)

// CompanyAxisID formats a company ID as the axis identifier used by rebalance jobs.
func CompanyAxisID(companyID int64) string {
	return strconv.FormatInt(companyID, 10)
}
