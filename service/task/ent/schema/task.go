package schema

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Task holds the schema definition for the Task entity.
type Task struct {
	ent.Schema
}

// Fields of the Queue.
func (Task) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New).Immutable(),
		field.Enum("status").
			Values(
				string(QueueStatusPending),
				string(QueueStatusRunning),
				string(QUeueStatusCompleted),
				string(QueueStatusFailed)).
			Default(string(QueueStatusPending)),
		field.JSON("body", json.RawMessage{}).Immutable(),
	}
}

type TaskStatus string

const (
	QueueStatusPending   TaskStatus = "pending"
	QueueStatusRunning   TaskStatus = "running"
	QUeueStatusCompleted TaskStatus = "completed"
	QueueStatusFailed    TaskStatus = "failed"
)

// Edges of the Queue.
func (Task) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tags", Tag.Type).Ref("tasks"),
	}
}
