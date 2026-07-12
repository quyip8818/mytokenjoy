package authz

import "context"

type RevisionReader interface {
	GetAuthzRevision(ctx context.Context, companyID int64) (int64, error)
}
