package task

import (
	"context"
	"slices"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"github.com/google/uuid"
	"rezics.com/task-queue/service/task/ent/task"
)

var (
	ErrTaskLockedByOther    = &errs.Error{Code: errs.FailedPrecondition, Message: "task is locked by another worker, you should not report it"}
	ErrWorkerAlreadyHasTask = &errs.Error{Code: errs.FailedPrecondition, Message: "worker is already holding a task"}
)

//encore:api auth method=POST path=/task/report tag:worker
func (s *Service) Report(ctx context.Context, req *ReportReq) error {
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

type ReportReq struct {
	WorkerID string      `header:"Worker-ID"`
	TaskID   string      `json:"id"`
	Status   task.Status `json:"status"`
}

var (
	ErrInvalidTaskID = &errs.Error{Code: errs.InvalidArgument, Message: "invalid task ID"}
	ErrInvalidStatus = &errs.Error{Code: errs.InvalidArgument, Message: "invalid status"}
)

func (q *ReportReq) Validate() error {
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
