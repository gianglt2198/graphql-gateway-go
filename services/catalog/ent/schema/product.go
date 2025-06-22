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

// Product holds the schema definition for the Product entity.
type Product struct {
	ent.Schema
}

// Fields of the Product.
func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("description").Nillable().Optional(),
		field.Float("price").Positive().Default(0),
		field.Int("stock").Positive().Default(0),
	}
}

func (Product) Mixin() []ent.Mixin {
	return []ent.Mixin{
		pnnid.MixinWithPrefix("prod"),
		mixin.AuthorMixin{},
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
		edge.To("category", Category.Type).Annotations(entgql.RelayConnection(), entsql.OnDelete(entsql.Cascade)),
	}
}
