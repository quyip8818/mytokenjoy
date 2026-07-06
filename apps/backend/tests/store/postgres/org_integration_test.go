//go:build integration

package postgres_test

import (
	"encoding/json"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestOrgIntegrationConnectedAndEnabledIndependent(t *testing.T) {
	st := testPostgresStore(t)
	ctx := testutil.Ctx()

	integration := types.OrgIntegration{
		Connected:      true,
		Enabled:        false,
		StartTime:      "03:00",
		FrequencyHours: 6,
	}
	if err := st.Org().SetIntegration(ctx, integration); err != nil {
		t.Fatal(err)
	}

	got, err := st.Org().Integration(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Connected || got.Enabled {
		t.Fatalf("expected connected=true enabled=false, got %+v", got)
	}
	if got.StartTime != "03:00" || got.FrequencyHours != 6 {
		t.Fatalf("unexpected sync config %+v", got)
	}
}

func TestIntegrationCredentialEncryptRoundTrip(t *testing.T) {
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	key := common.DevDefaultKey()
	payload, err := json.Marshal(types.FeishuCredential{
		Platform: types.PlatformFeishu, AppID: "cli_pg", AppSecret: "secret_pg",
	})
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := common.Encrypt(key, payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Org().SaveIntegrationCredential(ctx, types.PlatformFeishu, encrypted); err != nil {
		t.Fatal(err)
	}
	stored, err := st.Org().GetIntegrationCredential(ctx)
	if err != nil || stored == nil {
		t.Fatalf("expected stored credential, err=%v stored=%v", err, stored)
	}
	if stored.Platform != types.PlatformFeishu {
		t.Fatalf("unexpected platform %s", stored.Platform)
	}
}

func TestSaveIntegrationCredentialDoesNotOverwriteSyncConfig(t *testing.T) {
	st := testPostgresStore(t)
	ctx := testutil.Ctx()

	integration := types.OrgIntegration{
		Enabled:        true,
		StartTime:      "04:00",
		FrequencyHours: 12,
	}
	if err := st.Org().SetIntegration(ctx, integration); err != nil {
		t.Fatal(err)
	}

	key := common.DevDefaultKey()
	payload, err := json.Marshal(types.FeishuCredential{
		Platform: types.PlatformFeishu, AppID: "cli_pg", AppSecret: "secret_pg",
	})
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := common.Encrypt(key, payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Org().SaveIntegrationCredential(ctx, types.PlatformFeishu, encrypted); err != nil {
		t.Fatal(err)
	}

	got, err := st.Org().Integration(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Enabled || got.StartTime != "04:00" || got.FrequencyHours != 12 {
		t.Fatalf("sync config overwritten: %+v", got)
	}
	if got.Platform == nil || *got.Platform != types.PlatformFeishu {
		t.Fatalf("expected platform feishu, got %+v", got.Platform)
	}
	if len(got.EncryptedCredential) == 0 {
		t.Fatal("expected encrypted credential")
	}
}

func TestClearIntegrationCredentialSetsNull(t *testing.T) {
	st := testPostgresStore(t)
	ctx := testutil.Ctx()

	key := common.DevDefaultKey()
	payload, err := json.Marshal(types.FeishuCredential{
		Platform: types.PlatformFeishu, AppID: "cli_pg", AppSecret: "secret_pg",
	})
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := common.Encrypt(key, payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Org().SaveIntegrationCredential(ctx, types.PlatformFeishu, encrypted); err != nil {
		t.Fatal(err)
	}
	if err := st.Org().ClearIntegrationCredential(ctx); err != nil {
		t.Fatal(err)
	}

	got, err := st.Org().GetIntegrationCredential(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expected nil credential, got %+v", got)
	}
}

func TestFieldMappingsPersistRoundTrip(t *testing.T) {
	st := testPostgresStore(t)
	ctx := testutil.Ctx()

	mappings := []types.FieldMapping{
		{SourceField: "user_name", SourceLabel: "用户姓名", TargetField: "name", Required: true},
		{SourceField: "mobile", SourceLabel: "手机号码", TargetField: "phone", Required: true},
	}
	if err := st.Org().SetFieldMappings(ctx, mappings); err != nil {
		t.Fatal(err)
	}
	got, err := st.Org().FieldMappings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].TargetField != "name" || got[1].TargetField != "phone" {
		t.Fatalf("unexpected mappings %+v", got)
	}
}
