package auth

import (
	"context"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"github.com/google/uuid"
)

//encore:api auth method=DELETE path=/auth/cancel
func (s *Service) Cancel(ctx context.Context) error {
	uid, _ := auth.UserID()

	if err := s.Database.User.DeleteOneID(uuid.MustParse(string(uid))).Exec(ctx); err != nil {
		return &errs.Error{Code: errs.Internal}
	}

	return nil
}
