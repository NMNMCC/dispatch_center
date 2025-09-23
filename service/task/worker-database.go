package task

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"rezics.com/task-queue/service/task/ent/task"
	"rezics.com/task-queue/service/task/ent/worker"
)

var (
	ErrWorkerTaskMismatch = errors.New("task is not assigned to this worker")
)

func (s *Service) Init(ctx context.Context, uid, wid uuid.UUID, d time.Duration) error {
	if err := s.Database.Worker.Create().SetID(wid).SetEndOfLife(time.Now().Add(d)).Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Service) Check(ctx context.Context, uid, wid uuid.UUID) (tid uuid.UUID, locked bool, err error) {
	w, err := s.Database.Worker.Query().Where(worker.ID(wid)).Only(ctx)
	if err != nil {
		return uuid.Nil, false, err
	}

	if w.EndOfLife.Before(time.Now()) {
		s.Database.Worker.DeleteOne(w).ExecX(ctx)
		return uuid.Nil, false, nil
	}

	if locked, err := w.QueryTask().Exist(ctx); err != nil {
		return uuid.Nil, false, err
	} else {
		return w.ID, locked, nil
	}
}

func (s *Service) KeepAlive(ctx context.Context, uid, wid uuid.UUID, d time.Duration) error {
	w, err := s.Database.Worker.Query().Where(worker.ID(wid)).Only(ctx)
	if err != nil {
		return err
	}

	if err := w.Update().SetEndOfLife(time.Now().Add(d)).Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Service) Lock(ctx context.Context, uid, wid, tid uuid.UUID, d time.Duration) error {
	t, err := s.Database.Task.Query().Where(task.ID(tid)).Only(ctx)
	if err != nil {
		return err
	}

	w, err := s.Database.Worker.Query().Where(worker.ID(wid)).Only(ctx)
	if err != nil {
		return err
	}

	if err := t.Update().SetWorker(w).SetStatus(task.StatusRunning).Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Service) Unlock(ctx context.Context, uid, wid, tid uuid.UUID, d time.Duration) error {
	t, err := s.Database.Task.Query().Where(task.ID(tid)).Only(ctx)
	if err != nil {
		return err
	}

	w, err := s.Database.Worker.Query().Where(worker.ID(wid)).Only(ctx)
	if err != nil {
		return err
	}

	if t.Edges.Worker.ID != w.ID {
		return ErrWorkerTaskMismatch
	}

	if err := t.Update().ClearWorker().Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Service) Clean(ctx context.Context) error {
	now := time.Now()

	if err := s.Database.Task.Update().Where(task.HasWorkerWith(worker.EndOfLifeLT(now))).Exec(ctx); err != nil {
		return err
	}
	if _, err := s.Database.Worker.Delete().Where(worker.EndOfLifeLT(now)).Exec(ctx); err != nil {
		return err
	}

	return nil
}
