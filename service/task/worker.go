package task

import (
	"context"
	"encoding/json"
	"slices"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"rezics.com/task-queue/service/task/ent"
	"rezics.com/task-queue/service/task/ent/tag"
	"rezics.com/task-queue/service/task/ent/task"
)

var (
	ErrWorkerAlreadyExists = &errs.Error{Code: errs.AlreadyExists, Message: "worker already exists"}
)

//encore:api auth method=POST path=/task/worker/register tag:worker
func (s *Service) WorkerRegister(ctx context.Context) (*WorkerRegisterRes, error) {
	uid, _ := auth.UserID()

	wid := uuid.NewString()

	if err := s.Init(ctx, uuid.MustParse(string(uid)), uuid.MustParse(wid), cfg.WorkerKeepAliveDuration); err != nil {
		return nil, ErrWorkerAlreadyExists
	}

	return &WorkerRegisterRes{WorkerID: wid}, nil
}

type WorkerRegisterRes struct {
	WorkerID string `json:"worker_id"`
}

//encore:api auth method=POST path=/task/worker/keepalive tag:worker
func (s *Service) WorkerKeepAlive(ctx context.Context, req *WorkerKeepAliveReq) error {
	uid, _ := auth.UserID()

	if err := s.KeepAlive(ctx, uuid.MustParse(string(uid)), uuid.MustParse(req.WorkerID), cfg.WorkerKeepAliveDuration); err != nil {
		return ErrUnknown
	}

	return nil
}

type WorkerKeepAliveReq struct {
	WorkerID string `header:"Worker-ID"`
}

var (
	ErrTaskNextNotFound = &errs.Error{Code: errs.NotFound, Message: "no matching tasks found"}
)

//encore:api auth method=POST path=/task/worker/next tag:worker
func (s *Service) WorkerNext(
	ctx context.Context,
	req *WorkerNextReq) (*TaskRes, error) {
	uid, _ := auth.UserID()

	if tid, locked, err := s.Check(ctx,
		uuid.MustParse(string(uid)),
		uuid.MustParse(req.WorkerID)); err != nil || locked {
		return nil, &errs.Error{Code: errs.FailedPrecondition, Message: "worker is already processing a task: " + tid.String()}
	}

	t, err := s.Database.Task.Query().
		Where(task.HasTagsWith(tag.NameIn(req.Tags...)), task.Not(task.HasWorker())).
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
		ID:        t.ID.String(),
		Tags:      lo.Map(t.Edges.Tags, func(t *ent.Tag, _ int) string { return t.Name }),
		Body:      t.Body,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}, nil
}

type WorkerNextReq struct {
	WorkerID string   `header:"Worker-ID"`
	Tags     []string `json:"tags"`
}

type TaskRes struct {
	ID   string          `json:"id"`
	Tags []string        `json:"tags"`
	Body json.RawMessage `json:"task"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var (
	ErrTaskLockedByOther    = &errs.Error{Code: errs.FailedPrecondition, Message: "task is locked by another worker, you should not report it"}
	ErrWorkerAlreadyHasTask = &errs.Error{Code: errs.FailedPrecondition, Message: "worker is already holding a task"}
)

//encore:api auth method=POST path=/task/worker/report tag:worker
func (s *Service) WorkerReport(ctx context.Context, req *WorkerReportReq) error {
	uid, _ := auth.UserID()

	if tid, locked, err := s.Check(ctx, uuid.MustParse(string(uid)), uuid.MustParse(req.WorkerID)); err != nil {
		return ErrUnknown
	} else if locked || tid.String() != req.WorkerID {
		return ErrTaskLockedByOther
	}

	if err := s.Database.Task.UpdateOneID(uuid.MustParse(req.TaskID)).SetStatus(req.Status).Exec(ctx); err != nil {
		rlog.Error("failed to update task status", "error", err)
		return ErrUnknown
	}

	if err := s.Unlock(ctx, uuid.MustParse(string(uid)), uuid.MustParse(req.WorkerID), uuid.MustParse(req.TaskID), cfg.WorkerKeepAliveDuration); err != nil {
		rlog.Error("failed to unlock worker", "error", err)
		return ErrUnknown
	}

	return nil
}

type WorkerReportReq struct {
	WorkerID string      `header:"Worker-ID"`
	TaskID   string      `json:"id"`
	Status   task.Status `json:"status"`
}

var (
	ErrInvalidTaskID = &errs.Error{Code: errs.InvalidArgument, Message: "invalid task ID"}
	ErrInvalidStatus = &errs.Error{Code: errs.InvalidArgument, Message: "invalid status"}
)

func (q *WorkerReportReq) Validate() error {
	if _, err := uuid.Parse(q.TaskID); err != nil {
		return ErrInvalidTaskID
	}

	final := []task.Status{
		task.StatusCompleted,
		task.StatusFailed,
	}

	if !slices.Contains(final, q.Status) {
		return ErrInvalidStatus
	}

	return nil
}
