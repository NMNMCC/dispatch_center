package auth

import (
	"context"
	"encoding/hex"
	"slices"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"github.com/google/uuid"
	"rezics.com/task-queue/service/auth/ent/schema"
)

func NewKey() string {
	return hex.EncodeToString([]byte(uuid.NewString()))
}

var (
	ErrUnknown = &errs.Error{Code: errs.Internal, Message: "unknown error"}
)

//encore:api auth method=POST path=/auth/key/create
func (s *Service) CreateKey(ctx context.Context, req *CreateKeyReq) (*CreateKeyRes, error) {
	data, _ := auth.Data().(*AuthData)
	if !slices.Contains(data.Key.Permissions, string(schema.KeyPermissionCreateKey)) {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "create:key permission required"}
	}

	uid, _ := auth.UserID()

	tx, err := s.Database.Tx(ctx)
	if err != nil {
		return nil, ErrUnknown
	}

	key, err := tx.Key.
		Create().
		SetBody(NewKey()).
		SetPermissions(req.Permissions).
		SetRevokedAt(req.RevokedAt).
		SetUserID(uuid.MustParse(string(uid))).
		Save(ctx)
	if err != nil {
		tx.Rollback()
		return nil, ErrUnknown
	}

	return &CreateKeyRes{Key: key.Body}, nil
}

type CreateKeyReq struct {
	Permissions []string  `json:"permission"`
	RevokedAt   time.Time `json:"revoked_at"`
}

type CreateKeyRes struct {
	Key string `json:"key"`
}
