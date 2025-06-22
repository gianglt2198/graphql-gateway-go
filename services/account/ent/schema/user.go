package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gianglt2198/federation-go/package/modules/db/mixin"
	"github.com/gianglt2198/federation-go/package/modules/db/pnnid"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").Unique(),
		field.String("email").Unique(),
		field.String("password").Sensitive(),
		field.String("first_name"),
		field.String("last_name"),
		field.String("phone"),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		pnnid.MixinWithPrefix("user"),
		mixin.AuthorMixin{},
		mixin.SoftDeleteMixin{},
	}
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.Mutations(),
		entgql.MultiOrder(),
		entgql.RelayConnection(),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{}
}
