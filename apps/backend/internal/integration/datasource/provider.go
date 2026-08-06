// Package datasource implements external org data providers (Feishu, DingTalk).
// The domain interfaces are defined in domain/types/datasource.go.
package datasource

import "github.com/tokenjoy/backend/internal/domain/types"

// Type aliases — keep unqualified names within integration/datasource.
type RemoteDepartment = types.RemoteDepartment
type RemoteMember = types.RemoteMember
type Provider = types.DataSourceProvider
type Factory = types.DataSourceFactory
