package testutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type testSchema struct {
	baseURL string
	schema  string
	url     string
}

var testSchemaByName sync.Map
var ingestTestMu sync.Mutex

func openTestSchema(t *testing.T) (baseURL, schemaURL string) {
	t.Helper()
	if v, ok := testSchemaByName.Load(t.Name()); ok {
		h := v.(*testSchema)
		return h.baseURL, h.url
	}
	baseURL = defaultTestDatabaseURL()
	if baseURL == "" {
		t.Fatal("DATABASE_URL required; run: pnpm start:postgres")
	}
	schema := newTestSchemaName()
	adminPool, err := pgxpool.New(context.Background(), baseURL)
	if err != nil {
		t.Fatal(err)
	}
	schemaSQL := pgx.Identifier{schema}.Sanitize()
	if _, err := adminPool.Exec(context.Background(), "CREATE SCHEMA "+schemaSQL); err != nil {
		adminPool.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = adminPool.Exec(context.Background(), "DROP SCHEMA IF EXISTS "+schemaSQL+" CASCADE")
		adminPool.Close()
		testSchemaByName.Delete(t.Name())
	})
	h := &testSchema{
		baseURL: baseURL,
		schema:  schema,
		url:     withSearchPath(baseURL, schema),
	}
	testSchemaByName.Store(t.Name(), h)
	return h.baseURL, h.url
}

func TestSchemaURL(t *testing.T) string {
	t.Helper()
	_, schemaURL := openTestSchema(t)
	return schemaURL
}

func newTestSchemaName() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return "test_" + hex.EncodeToString(b[:])
}

func withSearchPath(dbURL, schema string) string {
	u, err := url.Parse(dbURL)
	if err != nil {
		panic(fmt.Sprintf("parse database url: %v", err))
	}
	q := u.Query()
	q.Set("options", fmt.Sprintf("-c search_path=%s,public", schema))
	u.RawQuery = q.Encode()
	return u.String()
}
