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
	ID                   uuid.UUID
	Name                 string
	Type                 string
	Status               string
	RootDeptID           *uuid.UUID
	NewAPIWalletUserID   *int64
	NewAPIWalletUsername string
	AuthzRevision        int64
	BillingCurrency      string
	FIFOHeadLotID        *uuid.UUID
	WalletRemain         int64
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// ConfiguredNewAPIWalletUserID returns the NewAPI wallet user id when it is positive.
func ConfiguredNewAPIWalletUserID(co *Company) (int64, bool) {
	if co == nil || co.NewAPIWalletUserID == nil || *co.NewAPIWalletUserID <= 0 {
		return 0, false
	}
	return *co.NewAPIWalletUserID, true
}

type CompanyRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Company, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]Company, error)
	Create(ctx context.Context, company Company) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateNewAPIWalletUserID(ctx context.Context, id uuid.UUID, walletUserID int64) error
	UpdateRootDeptID(ctx context.Context, id uuid.UUID, rootDeptID uuid.UUID) error
	GetAuthzRevision(ctx context.Context, id uuid.UUID) (int64, error)
	BumpAuthzRevision(ctx context.Context, id uuid.UUID) (int64, error)
	List(ctx context.Context) ([]Company, error)
	LockForUpdate(ctx context.Context, id uuid.UUID) (*Company, error)
	ApplyWalletDelta(ctx context.Context, id uuid.UUID, delta int64, fifoHeadLotID *uuid.UUID) error
	SetWalletRemain(ctx context.Context, id uuid.UUID, walletRemain int64, fifoHeadLotID *uuid.UUID) error
}
