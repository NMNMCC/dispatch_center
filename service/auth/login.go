package auth

import (
	"context"

	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"rezics.com/task-queue/service/auth/ent/user"
)

//encore:api public method=POST path=/auth/login
func (s *Service) Login(ctx context.Context, req *LoginReq) (*LoginRes, error) {
	u, err := s.Database.User.Query().Where(user.EmailEQ(req.Email)).First(ctx)
	if err != nil {
		rlog.Error("failed to find user", "error", err)
		return nil, &errs.Error{
			Code: errs.Internal,
		}
	}

	hash, err := FromHash(u.Password)
	if err != nil {
		return nil, &errs.Error{
			Code: errs.Internal,
		}
	}

	ok, err := VerifyHash(req.Password, hash)
	if err != nil {
		return nil, &errs.Error{
			Code: errs.Internal,
		}
	}
	if !*ok {
		return nil, &errs.Error{
			Code: errs.Unauthenticated,
		}
	}

	t := NewToken()
	if err := s.Database.User.UpdateOne(u).AppendTokens([]string{t}).Exec(ctx); err != nil {
		return nil, &errs.Error{
			Code: errs.Internal,
		}
	}

	return &LoginRes{Token: t}, nil
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRes struct {
	Token string `json:"token"`
}
