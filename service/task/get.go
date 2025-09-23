package task

import (
	"context"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"rezics.com/task-queue/service/task/ent"
	"rezics.com/task-queue/service/task/ent/tag"
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
		ID:        task.ID.String(),
		Tags:      lo.Map(task.Edges.Tags, func(t *ent.Tag, _ int) string { return t.Name }),
		Body:      task.Body,
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
	}, nil
}

type GetReq struct {
	TaskID string `query:"tid"`
}

//encore:api auth method=GET path=/task/get/worker
func (s *Service) GetWorker(ctx context.Context, req *GetWorkerReq) (*WorkerRes, error) {
	worker, err := s.Database.Worker.Query().WithTask().Where(worker.ID(uuid.MustParse(req.WorkerID))).Only(ctx)
	if err != nil {
		return nil, ErrUnknown
	}

	var currentTask TaskRes
	if worker.Edges.Task != nil {
		t := worker.Edges.Task
		tags, err := s.Database.Tag.Query().Where(tag.HasTasksWith(task.ID(t.ID))).All(ctx)
		if err != nil {
			return nil, ErrUnknown
		}

		currentTask = TaskRes{
			ID:   t.ID.String(),
			Tags: lo.Map(tags, func(tag *ent.Tag, _ int) string { return tag.Name }),
			Body: t.Body,
		}
	}

	return &WorkerRes{
		ID:           worker.ID.String(),
		RegisteredAt: worker.RegisteredAt,
		Task:         currentTask,
	}, nil
}

type GetWorkerReq struct {
	WorkerID string `query:"wid"`
}
