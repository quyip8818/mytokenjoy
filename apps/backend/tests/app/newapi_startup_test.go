//go:build testhook

package app_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestAppNewDoesNotWriteNewAPI(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{
		User: newapi.User{ID: 1, Quota: 0},
		CreateUserFn: func(context.Context, newapi.CreateUserRequest) (newapi.User, error) {
			t.Fatal("app.New must not call CreateUser")
			return newapi.User{}, nil
		},
		CreateTokenFn: func(context.Context, newapi.CreateTokenRequest) (newapi.Token, error) {
			t.Fatal("app.New must not call CreateToken")
			return newapi.Token{}, nil
		},
		UpdateTokenFn: func(context.Context, newapi.UpdateTokenRequest) (newapi.Token, error) {
			t.Fatal("app.New must not call UpdateToken")
			return newapi.Token{}, nil
		},
		EnsureGroupFn: func(context.Context, string, string) error {
			t.Fatal("app.New must not call EnsureGroup")
			return nil
		},
		UpsertChannelFn: func(context.Context, newapi.UpsertChannelRequest) (newapi.Channel, error) {
			t.Fatal("app.New must not call UpsertChannel")
			return newapi.Channel{}, nil
		},
	}
	testutil.NewTestAppWithOptions(t, func(cfg *config.Config) {
		testutil.WithDeployEnv("local")(cfg)
		testutil.WithNewAPIEnabled(true)(cfg)
	}, app.WithoutWorker(), app.WithAdminClient(stub))
}
