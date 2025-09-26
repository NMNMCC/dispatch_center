package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/go-playground/validator/v10"
	"rezics.com/task-queue/internal/util"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", util.NewUUIDv7()).Default(util.NewUUIDv7).Immutable(),
		field.String("email").Unique().NotEmpty().Validate(func(s string) error {
			return validator.New().Var(s, "required,email")
		}),
		field.String("password").Sensitive().NotEmpty(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("keys", Key.Type).Ref("user"),
	}
}
