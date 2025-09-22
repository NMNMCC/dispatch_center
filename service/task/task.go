package queue

import (
	"context"
	"encoding/json"

	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"rezics.com/task-queue/service/task/ent"
	"rezics.com/task-queue/service/task/ent/tag"

	"github.com/samber/lo"
)

//encore:api auth method=POST path=/task
func Enqueue(
	ctx context.Context,
	req *EnqueueRes) error {
	task, err := Database.Task.Create().
		SetBody(req.Body).
		Save(ctx)
	if err != nil {
		rlog.Error("failed to create task", "error", err)
		return &errs.Error{
			Code:    errs.Internal,
			Message: "failed to create task",
		}
	}
	for _, tag := range req.Tags {
		_, err := Database.Tag.Create().
			SetName(tag).
			AddTasks(task).
			Save(ctx)
		if err != nil {
			rlog.Error("failed to create tag", "error", err)
			return &errs.Error{
				Code:    errs.Internal,
				Message: "failed to create tag",
			}
		}
	}

	return nil
}

type EnqueueRes struct {
	Tags []string        `json:"tags"`
	Body json.RawMessage `json:"body"`
}

//encore:api auth method=GET path=/task
func Dequeue(
	ctx context.Context,
	res *DequeueRes) (*DequeueRep, error) {
	var (
		tags []*ent.Tag
		err  error
	)
	if len(res.Tags) == 0 {
		tags, err = Database.Tag.Query().All(ctx)
	} else {
		tags, err = Database.Tag.Query().
			Where(tag.NameIn(res.Tags...)).
			All(ctx)
	}
	if err != nil {
		rlog.Error("failed to get tasks", "error", err)
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to get tasks",
		}
	}

	tasks := make([]Task, 0)
	for _, tag := range tags {
		for _, task := range tag.Edges.Tasks {
			tasks = append(tasks, Task{
				Tags: lo.Map(task.Edges.Tags, func(tag *ent.Tag, _ int) string {
					return tag.Name
				}),
				Body: task.Body,
			})
		}
	}

	return &DequeueRep{
		Tasks: tasks,
	}, nil
}

type DequeueRes struct {
	Tags []string `query:"tags"`
}

type Task struct {
	Tags []string        `json:"tags"`
	Body json.RawMessage `json:"task"`
}

type DequeueRep struct {
	Tasks []Task `json:"tasks"`
}
