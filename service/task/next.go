package task

import (
	"context"
	"encoding/json"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"rezics.com/task-queue/service/task/ent"
	"rezics.com/task-queue/service/task/ent/predicate"
	"rezics.com/task-queue/service/task/ent/tag"
	"rezics.com/task-queue/service/task/ent/task"
)

var (
	ErrTaskNextNotFound = &errs.Error{Code: errs.NotFound, Message: "no matching tasks found"}
)

//encore:api auth method=POST path=/task/next tag:worker
func (s *Service) Next(
	ctx context.Context,
	req *NextReq) (*TaskRes, error) {
	uid, _ := auth.UserID()

	if tid, locked, err := s.Check(ctx,
		uuid.MustParse(string(uid)),
		uuid.MustParse(req.WorkerID)); err != nil || locked {
		return nil, &errs.Error{Code: errs.FailedPrecondition, Message: "worker is already processing a task: " + tid.String()}
	}

	where := []predicate.Task{task.HasTagsWith(tag.NameIn(req.Tags...))}

	t, err := s.Database.Task.Query().
		Where(where...).
		WithTags().
		Order(task.ByCreatedAt(sql.OrderAsc())).
		First(ctx)
	if err != nil {
		rlog.Error("failed to get tasks", "error", err)

		switch err.(type) {
		case *ent.NotFoundError:
			return nil, ErrTaskNextNotFound
		default:
			return nil, ErrUnknown
		}
	}

	tx, err := s.Database.Tx(ctx)
	if err != nil {
		rlog.Error("failed to start transaction", "error", err)
		return nil, ErrUnknown
	}

	if err := tx.Task.UpdateOne(t).SetStatus(task.StatusRunning).Exec(ctx); err != nil {
		tx.Rollback()
		rlog.Error("failed to update task status", "error", err)
		return nil, ErrUnknown
	}

	if err := s.Lock(ctx, uuid.MustParse(string(uid)), uuid.MustParse(req.WorkerID), t.ID, cfg.WorkerKeepAliveDuration); err != nil {
		tx.Rollback()
		rlog.Error("failed to lock worker", "error", err)
		return nil, ErrUnknown
	}

	tx.Commit()

	return &TaskRes{
		ID:   t.ID.String(),
		Tags: lo.Map(t.Edges.Tags, func(t *ent.Tag, _ int) string { return t.Name }),
		Body: t.Body,
	}, nil
}

type NextReq struct {
	WorkerID string   `header:"Worker-ID"`
	Tags     []string `json:"tags"`
}

type TaskRes struct {
	ID   string          `json:"id"`
	Tags []string        `json:"tags"`
	Body json.RawMessage `json:"task"`
}
