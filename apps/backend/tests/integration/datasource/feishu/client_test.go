package feishu_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource/feishu"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestClientUsesMockServer(t *testing.T) {
	t.Parallel()
	server := testutil.StartFeishuMockServer(t)
	client := feishu.NewClient(server.URL, types.FeishuCredential{
		Platform: types.PlatformFeishu, AppID: "cli_test", AppSecret: "secret",
	}, server.Client())

	if err := client.TestConnection(context.Background()); err != nil {
		t.Fatalf("test connection: %v", err)
	}
	depts, err := client.ListDepartments(context.Background())
	if err != nil || len(depts) != 1 {
		t.Fatalf("list departments: %v len=%d", err, len(depts))
	}
	members, failures, err := client.ListMembers(context.Background())
	if err != nil || len(members) != 1 || len(failures) != 0 {
		t.Fatalf("list members: %v members=%d failures=%d", err, len(members), len(failures))
	}
	member, err := client.SearchMember(context.Background(), "Mock")
	if err != nil || member.Name != "Mock User" {
		t.Fatalf("search member: %v %+v", err, member)
	}
}

func TestClientInvalidCredential(t *testing.T) {
	t.Parallel()
	server := testutil.StartFeishuAuthErrorServer(t)
	client := feishu.NewClient(server.URL, types.FeishuCredential{
		Platform: types.PlatformFeishu, AppID: "bad", AppSecret: "bad",
	}, server.Client())
	if err := client.TestConnection(context.Background()); err == nil {
		t.Fatal("expected auth error")
	}
}
