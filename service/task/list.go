package task

import (
	"context"
	"slices"
	"time"

	"encore.dev/beta/errs"
	"github.com/samber/lo"
	"rezics.com/task-queue/service/task/ent"
	"rezics.com/task-queue/service/task/ent/tag"
	"rezics.com/task-queue/service/task/ent/task"
)

//encore:api auth method=GET path=/task/list
func (s *Service) List(ctx context.Context, req *ListReq) (*ListRes, error) {
	ts, err := s.Database.Task.Query().
		WithTags().
		Where(
			task.CreatedAtGTE(req.After),
			task.CreatedAtLTE(req.Before),
			task.StatusIn(lo.Map(req.Status, func(s string, _ int) task.Status {
				return task.Status(s)
			})...),
			task.HasTagsWith(tag.NameIn(req.Tags...)),
		).
		Offset(req.Offset).
		Limit(req.Length).
		All(ctx)
	if err != nil {
		return nil, ErrUnknown
	}

	out := lo.Map(ts, func(t *ent.Task, _ int) TaskRes {
		return TaskRes{
			ID:   t.ID.String(),
			Tags: lo.Map(t.Edges.Tags, func(t *ent.Tag, _ int) string { return t.Name }),
			Body: t.Body,
		}
	})

	return &ListRes{Tasks: out}, nil
}

type ListReq struct {
	Tags   []string `query:"tags,omitempty"`
	Status []string `query:"status,omitempty"`

	Offset int `query:"offset,omitempty"`
	Length int `query:"length,omitempty"`

	Before time.Time `query:"before,omitempty"`
	After  time.Time `query:"after,omitempty"`
}

var (
	ErrLengthIsNotPositive = &errs.Error{Code: errs.InvalidArgument, Message: "length must be positive"}
	ErrInvalidTags         = &errs.Error{Code: errs.InvalidArgument, Message: "invalid tags"}
)

func (q *ListReq) Validate() error {
	if q.Before.IsZero() {
		q.Before = time.Now()
	}

	if q.Length <= 0 {
		return ErrLengthIsNotPositive
	} else if q.Length > 100 {
		q.Length = 100
	}

	if slices.Contains(q.Tags, "") {
		return ErrInvalidTags
	}

	all := []task.Status{
		task.StatusPending,
		task.StatusRunning,
		task.StatusCompleted,
		task.StatusFailed,
	}
	for _, status := range q.Status {
		if !slices.Contains(all, task.Status(status)) {
			return ErrInvalidStatus
		}
	}

	return nil
}

type ListRes struct {
	Tasks []TaskRes `json:"tasks"`
}
