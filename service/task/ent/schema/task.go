package schema

import (
	"encoding/json"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"rezics.com/task-queue/internal/util"
)

// Task holds the schema definition for the Task entity.
type Task struct {
	ent.Schema
}

// Fields of the Queue.
func (Task) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", util.NewUUIDv7()).Default(util.NewUUIDv7).Immutable(),
		field.Enum("status").
			Values(
				string(QueueStatusPending),
				string(QueueStatusRunning),
				string(QueueStatusCompleted),
				string(QueueStatusFailed)).
			Default(string(QueueStatusPending)),
		field.JSON("body", json.RawMessage{}).Immutable(),

		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

type TaskStatus string

const (
	QueueStatusPending   TaskStatus = "pending"
	QueueStatusRunning   TaskStatus = "running"
	QueueStatusCompleted TaskStatus = "completed"
	QueueStatusFailed    TaskStatus = "failed"
)

// Edges of the Queue.
func (Task) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tags", Tag.Type).Ref("tasks"),
		edge.To("worker", Worker.Type).Unique(),
	}
}
