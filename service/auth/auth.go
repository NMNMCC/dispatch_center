package auth

import (
	"context"

	"encore.dev/beta/auth"
)

// Data can be named whatever you prefer (but must be exported).
type Data struct {
	Username string
	// ...
}

// AuthHandler can be named whatever you prefer (but must be exported).
//
//encore:authhandler
func AuthHandler(ctx context.Context, token string) (auth.UID, *Data, error) {
	return auth.UID("test-id"), &Data{Username: "test"}, nil
}
