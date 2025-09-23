package auth

import (
	"context"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"rezics.com/task-queue/service/auth/ent/user"
)

type AuthData struct {
	Email string
}

//encore:authhandler
func (s *Service) AuthHandler(ctx context.Context, token string) (auth.UID, *AuthData, error) {
	u, err := s.Database.
		User.
		Query().
		Where(func(s *sql.Selector) { s.Where(sqljson.ValueContains(user.FieldTokens, token)) }).
		First(ctx)
	if err != nil {
		rlog.Error("failed to find user by token", "error", err)
		return "", nil, &errs.Error{
			Code: errs.Unauthenticated,
		}
	}

	return auth.UID(u.ID.String()), &AuthData{Email: u.Email}, nil
}
