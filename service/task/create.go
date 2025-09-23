package task

import (
	"context"
	"encoding/json"

	"encore.dev/rlog"
	"rezics.com/task-queue/service/task/ent/tag"
)

//encore:api auth method=POST path=/task/create
func (s *Service) Create(
	ctx context.Context,
	req *CreateReq) error {
	tx, err := s.Database.Tx(ctx)
	if err != nil {
		rlog.Error("failed to start transaction", "error", err)
		return ErrUnknown
	}

	t, err := tx.Task.Create().
		SetBody(req.Body).
		Save(ctx)
	if err != nil {
		rlog.Error("failed to create task", "error", err)
		_ = tx.Rollback()
		return ErrUnknown
	}

	for _, name := range req.Tags {
		if err := tx.Tag.Create().
			SetName(name).
			OnConflictColumns(tag.FieldName).
			Ignore().
			Exec(ctx); err != nil {
			tx.Rollback()

			rlog.Error("failed to upsert tag", "error", err, "tag", name)
			return ErrUnknown
		}

		if err := tx.Tag.Update().
			Where(tag.Name(name)).
			AddTasks(t).
			Exec(ctx); err != nil {
			tx.Rollback()

			rlog.Error("failed to attach task to tag", "error", err, "tag", name)
			return ErrUnknown
		}
	}

	if err := tx.Commit(); err != nil {
		rlog.Error("failed to commit transaction", "error", err)
		return ErrUnknown
	}

	return nil
}

type CreateReq struct {
	Tags []string        `json:"tags"`
	Body json.RawMessage `json:"body"`
}
