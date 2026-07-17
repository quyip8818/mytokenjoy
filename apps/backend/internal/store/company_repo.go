package store

import (
	"context"
	"time"
)

const (
	CompanyStatusActive    = "active"
	CompanyStatusSuspended = "suspended"

	CompanyTypeStandard   = "standard"
	CompanyTypeTrial      = "trial"
	CompanyTypeDemo       = "demo"
	CompanyTypeSelfhosted = "selfhosted"
	CompanyTypeTesting    = "testing"
)

type Company struct {
	ID                 int64
	Name               string
	Type               string
	Status             string
	RootDeptID         *string
	NewAPIWalletUserID *int64
	AuthzRevision      int64
	BillingCurrency    string
	FIFOHeadLotID      *string
	WalletRemain       float64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// ConfiguredNewAPIWalletUserID returns the NewAPI wallet user id when it is positive.
func ConfiguredNewAPIWalletUserID(co *Company) (int64, bool) {
	if co == nil || co.NewAPIWalletUserID == nil || *co.NewAPIWalletUserID <= 0 {
		return 0, false
	}
	return *co.NewAPIWalletUserID, true
}

type CompanyRepository interface {
	GetByID(ctx context.Context, id int64) (*Company, error)
	Create(ctx context.Context, company Company) error
	UpdateStatus(ctx context.Context, id int64, status string) error
	UpdateNewAPIWalletUserID(ctx context.Context, id int64, walletUserID int64) error
	UpdateRootDeptID(ctx context.Context, id int64, rootDeptID string) error
	GetAuthzRevision(ctx context.Context, id int64) (int64, error)
	BumpAuthzRevision(ctx context.Context, id int64) (int64, error)
	List(ctx context.Context) ([]Company, error)
	LockForUpdate(ctx context.Context, id int64) (*Company, error)
	ApplyWalletDelta(ctx context.Context, id int64, delta float64, fifoHeadLotID *string) error
	SetWalletRemain(ctx context.Context, id int64, walletRemain float64, fifoHeadLotID *string) error
}
