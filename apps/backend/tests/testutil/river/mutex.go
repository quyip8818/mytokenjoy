//go:build testhook

package riverfix

import "sync"

// TestMu serializes River client start/drain across packages. Parallel NOTIFY
// listeners otherwise steal jobs or leave retryable work that flakes drain tests.
var TestMu sync.Mutex
