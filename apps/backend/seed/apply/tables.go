package apply

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tokenjoy/backend/internal/store"
)

type TableWriter interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func ApplyTables(ctx context.Context, exec TableWriter, snap store.Snapshot) error {
	if err := insertSeedCurrencies(ctx, exec); err != nil {
		return err
	}
	if err := insertSeedCompany(ctx, exec, snap); err != nil {
		return err
	}
	tid := snap.Company.ID
	if err := insertSeedPermissions(ctx, exec, snap.Permissions); err != nil {
		return err
	}
	if err := insertSeedRoles(ctx, exec, tid, snap.Roles); err != nil {
		return err
	}
	roleIDByName := buildSeedRoleNameIndex(snap.Roles)
	if err := insertSeedModels(ctx, exec, tid, snap.Models); err != nil {
		return err
	}
	if err := insertSeedOrgNodes(ctx, exec, tid, snap.OrgNodes); err != nil {
		return err
	}
	if err := insertSeedMembers(ctx, exec, tid, snap.Members, roleIDByName); err != nil {
		return err
	}
	if err := insertSeedOrgIntegration(ctx, exec, tid, snap); err != nil {
		return err
	}
	if err := insertSeedBudget(ctx, exec, tid, snap); err != nil {
		return err
	}
	if err := insertSeedBudgetConsumed(ctx, exec, tid, snap); err != nil {
		return err
	}
	if err := insertSeedModelAllowlist(ctx, exec, tid, snap.ModelAllowlist); err != nil {
		return err
	}
	if err := insertSeedKeys(ctx, exec, tid, snap); err != nil {
		return err
	}
	if err := insertSeedAudit(ctx, exec, tid, snap); err != nil {
		return err
	}
	return nil
}
