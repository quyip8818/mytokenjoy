package store

import (
	"context"
	"time"

	"github.com/google/uuid"
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
	ID                    uuid.UUID
	Name                  string
	Industry              string
	Size                  string
	Type                  string
	Status                string
	RootDeptID            *uuid.UUID
	NewAPIWalletCompanyID *int64
	AuthzRevision         int64
	BillingCurrency       string
	FIFOHeadLotID         *uuid.UUID
	WalletQuotaRemain     int64
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// ConfiguredNewAPIWalletCompanyID returns the NewAPI wallet company id when it is positive.
func ConfiguredNewAPIWalletCompanyID(co *Company) (int64, bool) {
	if co == nil || co.NewAPIWalletCompanyID == nil || *co.NewAPIWalletCompanyID <= 0 {
		return 0, false
	}
	return *co.NewAPIWalletCompanyID, true
}

type CompanyRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Company, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]Company, error)
	Create(ctx context.Context, company Company) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateNewAPIWalletCompanyID(ctx context.Context, id uuid.UUID, walletCompanyID int64) error
	UpdateRootDeptID(ctx context.Context, id uuid.UUID, rootDeptID uuid.UUID) error
	GetAuthzRevision(ctx context.Context, id uuid.UUID) (int64, error)
	BumpAuthzRevision(ctx context.Context, id uuid.UUID) (int64, error)
	List(ctx context.Context) ([]Company, error)
	LockForUpdate(ctx context.Context, id uuid.UUID) (*Company, error)
	ApplyWalletDelta(ctx context.Context, id uuid.UUID, delta int64, fifoHeadLotID *uuid.UUID) error
	SetWalletQuotaRemain(ctx context.Context, id uuid.UUID, remain int64, fifoHeadLotID *uuid.UUID) error
}
