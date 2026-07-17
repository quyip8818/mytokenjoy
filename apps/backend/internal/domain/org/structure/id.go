package structure

import "github.com/google/uuid"

func generateID() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}
