package org_test

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/org/remote"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestUpdateSyncConfigValidation(t *testing.T) {
	t.Parallel()
	// Validation rejects before any store access, so nil store is safe here.
	deps := core.NewDeps(
		testutil.TestConfig(),
		nil, // store not reached for invalid inputs
		nil, nil, nil,
		common.NewDelayer(false),
		slog.Default(),
		permission.NewGrantNormalizer(),
	)
	svc := remote.New(deps, nil)

	tests := []struct {
		name    string
		cfg     types.SyncConfig
		wantErr string
	}{
		{
			name:    "negative FrequencyHours",
			cfg:     types.SyncConfig{FrequencyHours: -1, DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 10},
			wantErr: "frequencyHours",
		},
		{
			name:    "zero FrequencyHours",
			cfg:     types.SyncConfig{FrequencyHours: 0, DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 10},
			wantErr: "frequencyHours",
		},
		{
			name:    "negative DeleteMemberThreshold",
			cfg:     types.SyncConfig{FrequencyHours: 24, DeleteMemberThreshold: -1, DeleteDepartmentThreshold: 10},
			wantErr: "deleteMemberThreshold",
		},
		{
			name:    "negative DeleteDepartmentThreshold",
			cfg:     types.SyncConfig{FrequencyHours: 24, DeleteMemberThreshold: 10, DeleteDepartmentThreshold: -1},
			wantErr: "deleteDepartmentThreshold",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.UpdateSyncConfig(context.Background(), tc.cfg)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}
