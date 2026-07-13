//go:build testhook

package testutil

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func NewTestApp(t *testing.T, mutate func(*config.Config)) *app.App {
	return NewTestAppWithOptions(t, mutate, app.WithoutWorker(), app.WithAdminClient(defaultStubAdminClient()))
}

func NewTestAppWithOptions(t *testing.T, mutate func(*config.Config), opts ...app.Option) *app.App {
	t.Helper()
	cfg := TestConfig()
	if mutate != nil {
		mutate(&cfg)
	}
	_, st := NewTestStore(t, WithConfig(cfg))
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	application, err := app.NewWithStore(cfg, logger, st, opts...)
	if err != nil {
		t.Fatalf("create app: %v", err)
	}
	return application
}

func DefaultStubAdminClient() *mock.StubAdminClient {
	return defaultStubAdminClient()
}

func defaultStubAdminClient() *mock.StubAdminClient {
	var nextUserID int64 = 200
	var nextTokenID int64 = 1000
	return &mock.StubAdminClient{
		User: newapi.User{ID: nextUserID, Quota: 0},
		CreateUserFn: func(_ context.Context, _ newapi.CreateUserRequest) (newapi.User, error) {
			nextUserID++
			return newapi.User{ID: nextUserID, Quota: 0}, nil
		},
		CreateTokenFn: func(_ context.Context, _ newapi.CreateTokenRequest) (newapi.Token, error) {
			nextTokenID++
			return newapi.Token{
				ID:          nextTokenID,
				Key:         fmt.Sprintf("sk-test-%d", nextTokenID),
				RemainQuota: 1000,
			}, nil
		},
		GetTokenFn: func(_ context.Context, tokenID int64) (newapi.Token, error) {
			return newapi.Token{ID: tokenID, Key: fmt.Sprintf("sk-test-%d", tokenID), RemainQuota: 1000}, nil
		},
		GetTokenKeyFn: func(_ context.Context, tokenID int64) (string, error) {
			return fmt.Sprintf("sk-test-%d", tokenID), nil
		},
	}
}

func NewTestRouter(t *testing.T) http.Handler {
	t.Helper()
	return NewTestApp(t, nil).Router
}
