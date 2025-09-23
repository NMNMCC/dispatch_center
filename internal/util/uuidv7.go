package util

import (
	"github.com/google/uuid"
	"github.com/samber/lo"
)

func NewUUIDv7() uuid.UUID { return lo.Must(uuid.NewV7()) }
