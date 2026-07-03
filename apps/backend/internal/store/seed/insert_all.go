package seed

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tokenjoy/backend/internal/store"
)

type tableWriter interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func ApplyTables(ctx context.Context, exec tableWriter, snap store.Snapshot) error {
	if err := insertCompany(ctx, exec, snap); err != nil {
		return err
	}
	tid := snap.Company.ID
	if err := insertPermissions(ctx, exec, snap.Permissions); err != nil {
		return err
	}
	if err := insertRoles(ctx, exec, tid, snap.Roles); err != nil {
		return err
	}
	roleIDByName := buildRoleNameIndex(snap.Roles)
	if err := insertOrgNodes(ctx, exec, tid, snap.OrgNodes); err != nil {
		return err
	}
	if err := insertMembers(ctx, exec, tid, snap.Members, roleIDByName); err != nil {
		return err
	}
	if err := insertOrgIntegration(ctx, exec, tid, snap); err != nil {
		return err
	}
	if err := insertBudget(ctx, exec, tid, snap); err != nil {
		return err
	}
	if err := insertModels(ctx, exec, tid, snap.Models); err != nil {
		return err
	}
	if err := insertModelAllowlist(ctx, exec, tid, snap.ModelAllowlist); err != nil {
		return err
	}
	if err := insertKeys(ctx, exec, tid, snap); err != nil {
		return err
	}
	if err := insertAudit(ctx, exec, tid, snap); err != nil {
		return err
	}
	return nil
}
