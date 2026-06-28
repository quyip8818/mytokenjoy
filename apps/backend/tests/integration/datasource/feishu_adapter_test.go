package datasource_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestFeishuAdapterListDepartments(t *testing.T) {
	cfg := testutil.TestConfig()
	server := testutil.StartFeishuMockServer(t)
	cfg.FeishuBaseURL = server.URL
	factory := datasource.NewFactory(cfg)
	provider, err := factory.ForPlatform(types.PlatformFeishu, types.Credential{
		Platform: types.PlatformFeishu,
		Feishu: &types.FeishuCredential{
			Platform: types.PlatformFeishu, AppID: "cli_test", AppSecret: "secret_test",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	depts, err := provider.ListDepartments(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(depts) == 0 || depts[0].ExternalID != "od-1" {
		t.Fatalf("unexpected departments %+v", depts)
	}
}
