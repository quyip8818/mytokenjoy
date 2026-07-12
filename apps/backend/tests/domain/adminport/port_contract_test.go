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
	var _ adminport.Port = newapi.NewAdminPortAdapter(&newapi.Client{})
}

func TestAdminPortAdapterCreateTokenMapsResult(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{
		Token: newapi.Token{ID: 7, Key: "sk-demo", RemainQuota: 42},
	}
	port := newapi.NewAdminPortAdapter(stub)
	got, err := port.CreateToken(context.Background(), adminport.CreateTokenInput{Name: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 7 || got.Key != "sk-demo" || got.RemainQuota != 42 {
		t.Fatalf("unexpected token result: %+v", got)
	}
}
