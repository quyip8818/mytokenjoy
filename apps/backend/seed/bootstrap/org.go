package bootstrap

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"golang.org/x/crypto/bcrypt"
)

func insertRootOrg(ctx context.Context, exec TableWriter, companyID uuid.UUID, appCfg config.Config, cfg Config) error {
	rootID := deterministicOrgRootID(companyID)
	pathLabel := strings.ReplaceAll(rootID.String(), "-", "_")
	if _, err := exec.Exec(ctx, `
		INSERT INTO org_nodes (id, company_id, name, parent_id, path, type, sort_order)
		VALUES ($1, $2, $3, NULL, $4::ltree, 'dept', 0)
		ON CONFLICT (company_id, id) DO NOTHING
	`, rootID, companyID, cfg.Company.Name, pathLabel); err != nil {
		return fmt.Errorf("insert root org node: %w", err)
	}

	// Only create admin in prod mode. Demo/minimal modes have their own member data.
	if appCfg.BootstrapIsProd() && cfg.Admin.Email != "" {
		if err := insertAdminMember(ctx, exec, companyID, rootID, cfg); err != nil {
			return err
		}
	}
	return nil
}

func insertAdminMember(ctx context.Context, exec TableWriter, companyID, deptID uuid.UUID, cfg Config) error {
	now := time.Now().UTC()
	userID := deterministicAdminUserID(companyID)
	memberID := deterministicAdminMemberID(companyID)

	// Hash password if provided.
	var passwordHash *string
	if cfg.Admin.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(cfg.Admin.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash admin password: %w", err)
		}
		h := string(hash)
		passwordHash = &h
	}

	name := cfg.Admin.Name
	if name == "" {
		name = "管理员"
	}

	// 1. Create user row (holds auth identity: email + password).
	if _, err := exec.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, status, created_at, updated_at)
		VALUES ($1, $2, $3, 'active', $4, $4)
		ON CONFLICT (id) DO NOTHING
	`, userID, cfg.Admin.Email, passwordHash, now); err != nil {
		return fmt.Errorf("insert admin user: %w", err)
	}

	// 2. Create member row (company-scoped identity).
	if _, err := exec.Exec(ctx, `
		INSERT INTO members (id, company_id, user_id, name, department_id, status, source, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'active', 'bootstrap', $6, $6)
		ON CONFLICT (company_id, id) DO NOTHING
	`, memberID, companyID, userID, name, deptID, now); err != nil {
		return fmt.Errorf("insert admin member: %w", err)
	}

	// 3. Assign super admin role via member_roles.
	superAdminRoleID := grants.IDSuperAdmin
	if _, err := exec.Exec(ctx, `
		INSERT INTO member_roles (company_id, member_id, role_id)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`, companyID, memberID, superAdminRoleID); err != nil {
		return fmt.Errorf("insert admin member_role: %w", err)
	}

	return nil
}

func deterministicOrgRootID(companyID uuid.UUID) uuid.UUID {
	return uuid.NewSHA1(companyID, []byte("org-root"))
}

func deterministicAdminUserID(companyID uuid.UUID) uuid.UUID {
	return uuid.NewSHA1(companyID, []byte("bootstrap-admin-user"))
}

func deterministicAdminMemberID(companyID uuid.UUID) uuid.UUID {
	return uuid.NewSHA1(companyID, []byte("bootstrap-admin-member"))
}
