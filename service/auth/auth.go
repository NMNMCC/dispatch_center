package auth

import (
	"context"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"github.com/google/uuid"
	"rezics.com/task-queue/service/auth/ent/key"
)

//encore:authhandler
func (s *Service) AuthHandler(ctx context.Context, token string) (auth.UID, *AuthData, error) {
	k, err := s.Database.Key.Query().WithUser().Where(key.BodyEQ(token)).Only(ctx)
	if err != nil {
		rlog.Error("failed to find user by token", "error", err)
		return "", nil, &errs.Error{
			Code: errs.Unauthenticated,
		}
	}

	return auth.UID(k.Edges.User.ID.String()), &AuthData{Email: k.Edges.User.Email, Key: &KeyRes{
		ID:          k.ID,
		Body:        k.Body,
		Permissions: k.Permissions,
		CreatedAt:   k.CreatedAt,
		RevokedAt:   k.RevokedAt,
	}}, nil
}

type AuthData struct {
	Email string
	Key   *KeyRes
}

type KeyRes struct {
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// Body holds the value of the "body" field.
	Body string `json:"body,omitempty"`
	// Permissions holds the value of the "permissions" field.
	Permissions []string `json:"permissions,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// RevokedAt holds the value of the "revoked_at" field.
	RevokedAt time.Time `json:"revoked_at,omitempty"`
}
