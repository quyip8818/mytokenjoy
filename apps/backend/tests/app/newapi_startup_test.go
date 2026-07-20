//go:build testhook

package app_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestAppNewDoesNotWriteNewAPI(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{
		User: newapi.User{ID: 1, Quota: 0},
		CreateUserFn: func(context.Context, adminport.CreateUserInput) (adminport.UserResult, error) {
			t.Fatal("app.New must not call CreateUser")
			return adminport.UserResult{}, nil
		},
		CreateTokenFn: func(context.Context, adminport.CreateTokenInput) (adminport.TokenResult, error) {
			t.Fatal("app.New must not call CreateToken")
			return adminport.TokenResult{}, nil
		},
		UpdateTokenFn: func(context.Context, adminport.UpdateTokenInput) (adminport.TokenResult, error) {
			t.Fatal("app.New must not call UpdateToken")
			return adminport.TokenResult{}, nil
		},
		EnsureGroupFn: func(context.Context, string, string) error {
			t.Fatal("app.New must not call EnsureGroup")
			return nil
		},
		UpsertChannelFn: func(context.Context, adminport.UpsertChannelInput) (adminport.ChannelResult, error) {
			t.Fatal("app.New must not call UpsertChannel")
			return adminport.ChannelResult{}, nil
		},
	}
	testutil.NewTestAppWithOptions(t, func(cfg *config.Config) {
		testutil.WithDeployEnv("local")(cfg)
		testutil.WithNewAPIEnabled(true)(cfg)
	}, app.WithoutWorker(), app.WithAdminPort(stub))
}
