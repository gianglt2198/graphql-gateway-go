package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gianglt2198/federation-go/package/modules/db/pnnid"
)

type Session struct {
	ent.Schema
}

func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").Unique(),
		field.Time("last_used_at"),
	}
}

func (Session) Mixin() []ent.Mixin {
	return []ent.Mixin{
		pnnid.MixinWithPrefix("sess"),
	}
}

func (Session) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.Mutations(),
		entgql.MultiOrder(),
		entgql.RelayConnection(),
	}
}
