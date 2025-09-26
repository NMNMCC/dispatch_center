package auth

import (
	"context"
	"time"

	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"rezics.com/task-queue/service/auth/ent/schema"
	"rezics.com/task-queue/service/auth/ent/user"
)

//encore:api public method=POST path=/auth/login
func (s *Service) Login(ctx context.Context, req *LoginReq) (*LoginRes, error) {
	tx, err := s.Database.Tx(ctx)
	if err != nil {
		return nil, ErrUnknown
	}

	u, err := tx.User.Query().Where(user.EmailEQ(req.Email)).First(ctx)
	if err != nil {
		rlog.Error("failed to find user", "error", err)
		tx.Rollback()
		return nil, &errs.Error{
			Code: errs.Internal,
		}
	}

	hash, err := FromHash(u.Password)
	if err != nil {
		tx.Rollback()
		return nil, &errs.Error{
			Code: errs.Internal,
		}
	}

	ok, err := VerifyHash(req.Password, hash)
	if err != nil {
		tx.Rollback()
		return nil, &errs.Error{
			Code: errs.Internal,
		}
	}
	if !*ok {
		tx.Rollback()
		return nil, &errs.Error{
			Code: errs.Unauthenticated,
		}
	}

	t := NewKey()
	if err := tx.Key.Create().SetUser(u).SetBody(t).SetPermissions([]string{
		string(schema.KeyPermissionCreateKey),
		string(schema.KeyPermissionCreateTask),
		string(schema.KeyPermissionReadTask),
		string(schema.KeyPermissionUpdateTask),
		string(schema.KeyPermissionDeleteTask),
	}).SetRevokedAt(time.Now().Add(time.Hour)).Exec(ctx); err != nil {
		tx.Rollback()
		return nil, ErrUnknown
	}

	tx.Commit()

	return &LoginRes{Token: t}, nil
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRes struct {
	Token string `json:"token"`
}
