package schema

import (
	"errors"
	"slices"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"rezics.com/task-queue/internal/util"
)

// Key holds the schema definition for the Key entity.
type Key struct {
	ent.Schema
}

var (
	ErrInvalidPermission = errors.New("invalid permission")
)

// Fields of the Key.
func (Key) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", util.NewUUIDv7()).Default(util.NewUUIDv7).Immutable(),
		field.String("body").Immutable(),
		field.Strings("permissions").Default([]string{}).Immutable().Validate(func(s []string) error {
			all := []KeyPermission{
				KeyPermissionCreateTask,
				KeyPermissionReadTask,
				KeyPermissionUpdateTask,
				KeyPermissionDeleteTask,
				KeyPermissionKey,
				KeyPermissionWorker,
			}

			for _, p := range s {
				if !slices.Contains(all, KeyPermission(p)) {
					return ErrInvalidPermission
				}
			}

			return nil
		}),

		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("revoked_at"),
	}
}

type KeyPermission string

const (
	KeyPermissionCreateTask KeyPermission = "create:task"
	KeyPermissionReadTask   KeyPermission = "read:task"
	KeyPermissionUpdateTask KeyPermission = "update:task"
	KeyPermissionDeleteTask KeyPermission = "delete:task"
	KeyPermissionKey        KeyPermission = "key"
	KeyPermissionWorker     KeyPermission = "worker"
)

// Edges of the Key.
func (Key) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).Unique().Required().Immutable(),
	}
}
