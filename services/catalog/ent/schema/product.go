package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gianglt2198/graphql-gateway-go/pkg/modules/db/mixin"
	"github.com/gianglt2198/graphql-gateway-go/pkg/modules/db/pnnid"
)

type Product struct {
	ent.Schema
}

func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.Float("price"),
	}
}

func (Product) Mixin() []ent.Mixin {
	return []ent.Mixin{
		pnnid.MixinWithPrefix("prd"),
		mixin.AuthorMixin{},
	}
}

func (Product) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.Mutations(),
		entgql.MultiOrder(),
	}
}
