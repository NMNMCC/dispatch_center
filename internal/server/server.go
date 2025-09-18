package server

import (
	"encoding/json"
	"time"

	"github.com/labstack/echo/v4"
	"nmnm.cc/dispatch_center/internal/database"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

func (ts TaskStatus) IsValid() bool {
	switch ts {
	case TaskStatusPending, TaskStatusRunning, TaskStatusCompleted, TaskStatusFailed:
		return true
	}
	return false
}

type TaskBody struct {
	ID      string          `json:"id"`
	Status  TaskStatus      `json:"status"`
	Payload json.RawMessage `json:"payload"`
}

type Task struct {
	Locker string

	TaskBody
}

type Node struct {
	ID            string
	LastKeepalive time.Time
}

type Data struct {
	Task map[
	// Task Type
	string]map[
	// Task ID
	string]Task
	Node map[string]Node
}

type Server struct {
	e                    *echo.Echo
	d                    *database.DB[Data]
	nodeKeepaliveTimeout time.Duration
}

func (s *Server) Start(addr string) error {
	return s.e.Start(addr)
}

func (s *Server) startJanitor() {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			s.d.Write(func(data *Data) error {
				now := time.Now()
				for nid, node := range data.Node {
					if now.Sub(node.LastKeepalive) > s.nodeKeepaliveTimeout {
						// node is dead now

						for ttype, tasks := range data.Task {
							for tid, task := range tasks {
								if task.Locker == nid {
									task.Status = TaskStatusFailed
									data.Task[ttype][tid] = task
								}
							}
						}

						delete(data.Node, nid)
					}
				}
				return nil
			})
		}
	}()
}

func New(databasePath string, nodeKeepaliveTimeout time.Duration) (*Server, error) {
	e := echo.New()
	d, err := database.Open[Data](databasePath)
	if err != nil {
		return nil, err
	}

	s := &Server{
		e:                    e,
		d:                    d,
		nodeKeepaliveTimeout: nodeKeepaliveTimeout,
	}

	// Ensure maps are initialized in the DB
	if err := s.d.Write(func(data *Data) error {
		if data.Task == nil {
			data.Task = make(map[string]map[string]Task)
		}
		if data.Node == nil {
			data.Node = make(map[string]Node)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	s.rNode()
	s.rTask()
	s.startJanitor()

	return s, nil
}
