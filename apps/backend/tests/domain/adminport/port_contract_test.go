package adminport_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestAdminPortAdapterImplementsPort(t *testing.T) {
	t.Parallel()
	var _ adminport.Port = &newapi.Client{}
}

func TestAdminPortAdapterCreateTokenMapsResult(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{
		Token: newapi.Token{ID: 7, UserID: 501, Key: "sk-demo", RemainQuota: 42, Group: "platform_shared"},
	}
	port := stub
	got, err := port.CreateToken(context.Background(), adminport.CreateTokenInput{UserID: 501, Name: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 7 || got.UserID != 501 || got.Key != "sk-demo" || got.RemainQuota != 42 || got.Group != "platform_shared" {
		t.Fatalf("unexpected token result: %+v", got)
	}
}
