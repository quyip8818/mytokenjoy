package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"sms/backend/internal/config"
	"sms/backend/internal/store"
)

func main() {
	cfg, _ := config.Load()
	pool, err := store.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	// 重置 schema（项目未上线，直接重建）
	if _, err := pool.Exec(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"); err != nil {
		log.Fatal("reset schema:", err)
	}

	// 应用 schema
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		log.Fatal("read schema.sql:", err)
	}
	if _, err := pool.Exec(ctx, string(schema)); err != nil {
		log.Fatal("apply schema:", err)
	}
	fmt.Println("[seed] schema applied")

	// 管理员
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
	adminID := uuid.Must(uuid.NewV7())
	_, _ = pool.Exec(ctx, `
		INSERT INTO users (id, username, password_hash, real_name, email, role)
		VALUES ($1, 'admin', $2, '系统管理员', 'admin@example.com', 'admin')
		ON CONFLICT (username) DO NOTHING`, adminID, string(hash))

	// 评估权重
	_, _ = pool.Exec(ctx, `
		INSERT INTO evaluation_weights (id, dimension, weight) VALUES
			($1, 'quality', 30),
			($2, 'performance', 20),
			($3, 'price', 20),
			($4, 'service', 20),
			($5, 'compliance', 10)
		ON CONFLICT (dimension) DO NOTHING`,
		uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()))

	// OAuth client for TokenJoy sync
	syncSecretHash, _ := bcrypt.GenerateFromPassword([]byte("e2e-test-secret"), 10)
	_, _ = pool.Exec(ctx, `
		INSERT INTO oauth_clients (id, client_id, client_secret_hash, scope)
		VALUES ($1, 'tokenjoy-sync', $2, 'sync:read')
		ON CONFLICT (client_id) DO NOTHING`,
		uuid.Must(uuid.NewV7()), string(syncSecretHash))

	fmt.Println("[seed] done. Login: admin / admin123")
}
