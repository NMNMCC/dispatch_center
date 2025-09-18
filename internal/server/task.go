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
		s.d.Read(func(data *Data) error {
			for _, task := range data.Task[ttype] {
				tasks = append(tasks, task.TaskBody)
			}

			return nil
		})

		return c.JSON(http.StatusOK, tasks)
	})

	s.e.GET("/task/:type/:task_id", func(c echo.Context) error {
		ttype := c.Param("type")
		tid := c.Param("task_id")

		var task TaskBody
		if err := s.d.Read(func(data *Data) error {
			_task, ok := data.Task[ttype][tid]
			if !ok {
				return c.NoContent(http.StatusNotFound)
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

		s.d.Write(func(data *Data) error {
			delete(data.Task[ttype], tid)
			return nil
		})

		return c.NoContent(http.StatusNoContent)
	})

	s.e.PUT("/task/:type", func(c echo.Context) error {
		ttype := c.Param("type")

		var payload json.RawMessage
		if err := c.Bind(&payload); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		id := uuid.NewString()
		s.d.Write(func(data *Data) error {
			data.Task[ttype][id] = Task{
				TaskBody: TaskBody{
					ID:      id,
					Status:  TaskStatusPending,
					Payload: payload,
				},
			}
			return nil
		})

		return c.String(http.StatusOK, id)
	})
}
