package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"rezics.com/task-queue/internal/util"
)

// Worker holds the schema definition for the Worker entity.
type Worker struct {
	ent.Schema
}

// Fields of the Worker.
func (Worker) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", util.NewUUIDv7()).Default(util.NewUUIDv7).Immutable(),
		field.Time("end_of_life"),

		field.Time("registered_at").Default(time.Now).Immutable(),
	}
}

// Edges of the Worker.
func (Worker) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("task", Task.Type).Ref("worker").
			Unique(),
	}
}
