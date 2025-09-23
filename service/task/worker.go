package task

import (
	"context"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"github.com/google/uuid"
)

var (
	ErrWorkerAlreadyExists = &errs.Error{Code: errs.AlreadyExists, Message: "worker already exists"}
)

//encore:api auth method=POST path=/worker/register
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

//encore:api auth method=POST path=/worker/keepalive
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
