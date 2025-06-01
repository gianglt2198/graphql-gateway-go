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

type Category struct {
	ent.Schema
}

func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.Float("description"),
		field.Enum("status").
			NamedValues(
				"Active", "ACTIVE",
				"Inactive", "INACTIVE",
			).
			Default("ACTIVE").
			Annotations(
				entgql.OrderField("STATUS"),
			),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Annotations(
				entgql.OrderField("CREATED_AT"),
			),
	}
}

func (Category) Mixin() []ent.Mixin {
	return []ent.Mixin{
		pnnid.MixinWithPrefix("ctg"),
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
		edge.From("categories_products", Product.Type).Ref("products_categories").
			Annotations(
				entgql.RelayConnection(),
			),
	}
}
