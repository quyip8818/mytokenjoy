package store

import (
	"context"
	"time"
)

const (
	CompanyStatusActive    = "active"
	CompanyStatusSuspended = "suspended"
)

type Company struct {
	ID                 int64
	Slug               string
	Name               string
	Status             string
	RootDeptID         *string
	NewAPIWalletUserID *int64
	PackageID          *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type CompanyRepository interface {
	GetByID(ctx context.Context, id int64) (*Company, error)
	GetBySlug(ctx context.Context, slug string) (*Company, error)
	Create(ctx context.Context, company Company) error
	UpdateStatus(ctx context.Context, id int64, status string) error
	UpdatePackageID(ctx context.Context, id int64, packageID *string) error
	UpdateNewAPIWalletUserID(ctx context.Context, id int64, walletUserID int64) error
	UpdateRootDeptID(ctx context.Context, id int64, rootDeptID string) error
	List(ctx context.Context) ([]Company, error)
}
