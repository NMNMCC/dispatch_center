package auth

import (
	"encoding/hex"

	"github.com/google/uuid"
)

func NewToken() string {
	return hex.EncodeToString([]byte(uuid.NewString()))
}
