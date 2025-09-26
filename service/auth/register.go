package auth

import (
	"context"

	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"github.com/go-playground/validator/v10"
)

//encore:api public method=POST path=/auth/register
func (s *Service) Register(ctx context.Context, req *RegisterReq) error {
	if h, err := NewHash(req.Password); err != nil {
		return &errs.Error{
			Code: errs.Internal,
		}
	} else {
		req.Password = h
	}

	if err := s.Database.
		User.
		Create().
		SetEmail(req.Email).
		SetPassword(req.Password).
		Exec(ctx); err != nil {
		rlog.Error("failed to create user", "error", err)
		return &errs.Error{
			Code: errs.Internal,
		}
	}

	return nil
}

type RegisterReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var (
	ErrInvalidEmail    = &errs.Error{Code: errs.InvalidArgument, Message: "invalid email"}
	ErrInvalidPassword = &errs.Error{Code: errs.InvalidArgument, Message: "invalid password"}
)

func (q *RegisterReq) Validate() error {
	v := validator.New()

	if err := v.Var(q.Email, "required,email"); err != nil {
		return ErrInvalidEmail
	}

	if err := v.Var(q.Password, "required,min=8"); err != nil {
		return ErrInvalidPassword
	}

	return nil
}
