package task

import (
	"context"

	"github.com/google/uuid"
)

//encore:api auth method=DELETE path=/task/delete
func (s *Service) Delete(ctx context.Context, req *DeleteReq) error {
	if err := s.Database.Task.DeleteOneID(uuid.MustParse(req.TaskID)).Exec(ctx); err != nil {
		return ErrUnknown
	}

	return nil
}

type DeleteReq struct {
	TaskID string `query:"tid"`
}
