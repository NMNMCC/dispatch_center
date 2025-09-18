package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (s *Server) rTask() {
	s.e.GET("/task/:type", func(c echo.Context) error {
		ttype := c.Param("type")

		var tasks []TaskBody
		if err := s.d.Read(func(data *Data) error {
			for _, task := range data.Task[ttype] {
				tasks = append(tasks, task.TaskBody)
			}

			return nil
		}); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, tasks)
	})

	s.e.GET("/task/:type/:task_id", func(c echo.Context) error {
		ttype := c.Param("type")
		tid := c.Param("task_id")

		var task TaskBody
		if err := s.d.Read(func(data *Data) error {
			typedTask, ok := data.Task[ttype]
			if !ok {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			_task, ok := typedTask[tid]
			if !ok {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			task = _task.TaskBody
			return nil
		}); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, task)
	})

	s.e.DELETE("/task/:type/:task_id", func(c echo.Context) error {
		ttype := c.Param("type")
		tid := c.Param("task_id")

		if err := s.d.Write(func(data *Data) error {
			if _, ok := data.Task[ttype]; !ok {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			delete(data.Task[ttype], tid)
			return nil
		}); err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	})

	s.e.PUT("/task/:type", func(c echo.Context) error {
		ttype := c.Param("type")

		var payload json.RawMessage
		if err := c.Bind(&payload); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		id := uuid.NewString()
		if err := s.d.Write(func(data *Data) error {
			if data.Task == nil {
				data.Task = make(map[string]map[string]Task)
			}
			if _, ok := data.Task[ttype]; !ok {
				data.Task[ttype] = make(map[string]Task)
			}
			data.Task[ttype][id] = Task{
				TaskBody: TaskBody{
					ID:      id,
					Status:  TaskStatusPending,
					Payload: payload,
				},
			}
			return nil
		}); err != nil {
			return err
		}

		return c.String(http.StatusOK, id)
	})
}
