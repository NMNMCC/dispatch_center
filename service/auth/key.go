package auth

import (
	"context"
	"encoding/hex"
	"slices"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/middleware"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"rezics.com/task-queue/service/auth/ent"
	"rezics.com/task-queue/service/auth/ent/key"
	"rezics.com/task-queue/service/auth/ent/schema"
	"rezics.com/task-queue/service/auth/ent/user"
)

func NewKey() string {
	return hex.EncodeToString([]byte(uuid.NewString()))
}

var (
	ErrUnknown = &errs.Error{Code: errs.Internal, Message: "unknown error"}
)

//encore:middleware target=tag:key
func (s *Service) KeyAuthMiddleware(req middleware.Request, next middleware.Next) middleware.Response {
	data, _ := auth.Data().(*AuthData)
	if !slices.Contains(data.Key.Permissions, string(schema.KeyPermissionKey)) {
		return middleware.Response{
			Err: &errs.Error{Code: errs.Unauthenticated, Message: "key permission required"},
		}
	}

	return next(req)
}

//encore:api auth method=POST path=/auth/key/create tag:key
func (s *Service) CreateKey(ctx context.Context, req *CreateKeyReq) (*CreateKeyRes, error) {
	if req.RevokedAt.Before(time.Now()) {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "revoked_at must be in the future",
		}
	}
	if len(req.Permissions) == 0 {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "at least one permission is required",
		}
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

//encore:api auth method=GET path=/auth/key/list tag:key
func (s *Service) ListKey(ctx context.Context) (*ListKeyRes, error) {
	data, _ := auth.Data().(*AuthData)

	keys, err := s.Database.Key.
		Query().
		Where(key.HasUserWith(user.EmailEQ(data.Email))).
		All(ctx)
	if err != nil {
		return nil, ErrUnknown
	}

	return &ListKeyRes{Keys: lo.Map(keys, func(k *ent.Key, _ int) *KeyRes {
		return &KeyRes{
			Body:        k.Body,
			Permissions: k.Permissions,
			CreatedAt:   k.CreatedAt,
			RevokedAt:   k.RevokedAt,
		}
	})}, nil
}

type ListKeyRes struct {
	Keys []*KeyRes `json:"key"`
}

//encore:api auth method=POST path=/auth/key/revoke tag:key
func (s *Service) RevokeKey(ctx context.Context, req *RevokeKeyReq) error {
	data, _ := auth.Data().(*AuthData)

	_, err := s.Database.Key.
		Update().
		Where(
			key.BodyEQ(req.Key),
			key.HasUserWith(user.EmailEQ(data.Email)),
		).
		SetRevokedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return ErrUnknown
	}

	return nil
}

type RevokeKeyReq struct {
	Key string `json:"key"`
}
