package structure

import (
	"crypto/rand"
	"fmt"
	"time"
)

func generateID(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%d-%x", prefix, time.Now().UnixMilli(), b)
}
