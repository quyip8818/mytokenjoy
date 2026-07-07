package common_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func TestResolveDemoMemberName(t *testing.T) {
	members := []types.Member{
		{ID: "m1", Name: "Alice"},
		{ID: "m2", Name: "Bob"},
	}

	tests := []struct {
		name     string
		memberID string
		want     string
	}{
		{"found member", "m1", "Alice"},
		{"found second member", "m2", "Bob"},
		{"empty memberID", "", "审批人"},
		{"not found", "m3", "审批人"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := common.ResolveDemoMemberName(tt.memberID, members)
			if got != tt.want {
				t.Errorf("ResolveDemoMemberName(%q) = %q, want %q", tt.memberID, got, tt.want)
			}
		})
	}
}
