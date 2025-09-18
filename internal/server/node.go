package server

import (
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (s *Server) rNode() {
	// node keepalive, otherwise task fails
	s.e.HEAD("/node/keepalive", func(c echo.Context) error {
		header := c.Request().Header
		nid := header.Get("NodeID")

		s.d.Write(func(data *Data) error {
			data.Node[nid] = Node{
				ID:            nid,
				LastKeepalive: time.Now(),
			}

			return nil
		})

		return c.NoContent(http.StatusOK)
	})

	// node put task report
	s.e.PUT("/node/report/:type/:task_id", func(c echo.Context) error {
		header := c.Request().Header
		nid := header.Get("NodeID")

		if err := s.d.Read(func(data *Data) error {
			if _, ok := data.Node[nid]; !ok {
				return echo.NewHTTPError(http.StatusForbidden, "node not registered")
			}
			return nil
		}); err != nil {
			return err
		}

		ttype := c.Param("type")
		tid := c.Param("task_id")
		body, err := io.ReadAll(c.Request().Body)
		status := TaskStatus(body)
		if !status.IsValid() {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid task status")
		}

		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if err := s.d.Write(func(data *Data) error {
			typedTask, ok := data.Task[ttype]
			if !ok {
				return echo.NewHTTPError(http.StatusNotFound)
			}

			task, ok := typedTask[tid]
			if !ok {
				return echo.NewHTTPError(http.StatusNotFound)
			}

			if task.Locker != nid {
				return echo.NewHTTPError(http.StatusForbidden)
			}

			task.Status = status
			typedTask[tid] = task
			data.Task[ttype] = typedTask

			return nil
		}); err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)
	})

	// node get next task
	s.e.GET("/node/next/:type", func(c echo.Context) error {
		header := c.Request().Header
		ttype := c.Param("type")
		nid := header.Get("NodeID")

		if err := s.d.Read(func(data *Data) error {
			if _, ok := data.Node[nid]; !ok {
				return echo.NewHTTPError(http.StatusForbidden, "node not registered")
			}
			return nil
		}); err != nil {
			return err
		}

		var taskToRun *Task
		var found bool

		if err := s.d.Write(func(data *Data) error {
			// check if node already has a task
			for _, task := range data.Task[ttype] {
				if task.Locker == nid {
					// The node is already working on a task, return the payload so it can continue
					taskToRun = &task
					found = true
					return nil
				}
			}

			// find a pending task and lock it
			for id, task := range data.Task[ttype] {
				if task.Status == TaskStatusPending {
					task.Locker = nid
					task.Status = TaskStatusInProgress
					data.Task[ttype][id] = task
					taskToRun = &task
					found = true
					break
				}
			}

			return nil
		}); err != nil {
			return err
		}

		if !found {
			return c.NoContent(http.StatusNotFound)
		}

		return c.JSON(http.StatusOK, taskToRun.Payload)
	})
}
