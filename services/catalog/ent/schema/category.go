package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/gianglt2198/federation-go/package/modules/db/mixin"
	"github.com/gianglt2198/federation-go/package/modules/db/pnnid"
)

type Category struct {
	ent.Schema
}

func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("description").Nillable().Optional(),
	}
}

func (Category) Mixin() []ent.Mixin {
	return []ent.Mixin{
		pnnid.MixinWithPrefix("cat"),
		mixin.AuthorMixin{},
	}
}

func (Category) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.Mutations(),
		entgql.MultiOrder(),
		entgql.RelayConnection(),
	}
}

func (Category) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("products", Product.Type).Ref("categories").
			Annotations(entgql.RelayConnection(), entsql.OnDelete(entsql.Cascade)),
	}
}
