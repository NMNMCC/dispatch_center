package auth

import (
	"context"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
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
	if k.RevokedAt.Before(time.Now()) {
		return "", nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "key revoked",
		}
	}

	return auth.UID(k.Edges.User.ID.String()), &AuthData{Email: k.Edges.User.Email, Key: &KeyRes{
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
	// Body holds the value of the "body" field.
	Body string `json:"body,omitempty"`
	// Permissions holds the value of the "permissions" field.
	Permissions []string `json:"permissions,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitzero"`
	// RevokedAt holds the value of the "revoked_at" field.
	RevokedAt time.Time `json:"revoked_at,omitzero"`
}
