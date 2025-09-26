package task

import (
	"context"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"rezics.com/task-queue/service/task/ent"
	"rezics.com/task-queue/service/task/ent/task"
	"rezics.com/task-queue/service/task/ent/worker"
)

//encore:api auth method=GET path=/task/get
func (s *Service) Get(ctx context.Context, req *GetReq) (*TaskRes, error) {
	task, err := s.Database.Task.Query().WithTags().Where(task.ID(uuid.MustParse(req.TaskID))).Only(ctx)
	if err != nil {
		return nil, ErrUnknown
	}

	return &TaskRes{
		ID:   task.ID.String(),
		Tags: lo.Map(task.Edges.Tags, func(t *ent.Tag, _ int) string { return t.Name }),
		Body: task.Body,
	}, nil
}

type GetReq struct {
	TaskID string `query:"tid"`
}

//encore:api auth method=GET path=/task/get/worker
func (s *Service) GetWorker(ctx context.Context, req *GetWorkerReq) (*WorkerRes, error) {
	raw_worker, err := s.Database.Worker.Query().WithTask().Where(worker.ID(uuid.MustParse(req.WorkerID))).Only(ctx)
	if err != nil {
		return nil, ErrUnknown
	}

	worker := &WorkerRes{
		ID:           raw_worker.ID.String(),
		RegisteredAt: raw_worker.RegisteredAt,
	}

	if raw_worker.Edges.Task != nil {
		worker.Task = &TaskRes{
			ID:        raw_worker.Edges.Task.ID.String(),
			Tags:      lo.Map(raw_worker.Edges.Task.Edges.Tags, func(t *ent.Tag, _ int) string { return t.Name }),
			Body:      raw_worker.Edges.Task.Body,
			CreatedAt: raw_worker.Edges.Task.CreatedAt,
			UpdatedAt: raw_worker.Edges.Task.UpdatedAt,
		}
	}

	return worker, nil
}

type GetWorkerReq struct {
	WorkerID string `query:"wid"`
}
