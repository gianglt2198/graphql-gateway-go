package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/gianglt2198/graphql-gateway-go/pkg/modules/db/pnnid"
)

type Product struct {
	ent.Schema
}

func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.Float("price"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Annotations(
				entgql.OrderField("CREATED_AT"),
			),
	}
}

func (Product) Mixin() []ent.Mixin {
	return []ent.Mixin{
		pnnid.MixinWithPrefix("prd"),
	}
}

func (Product) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.Mutations(),
		entgql.MultiOrder(),
		entgql.RelayConnection(),
	}
}

func (Product) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("products_categories", Category.Type).
			Annotations(
				entgql.RelayConnection(),
			),
	}
}
