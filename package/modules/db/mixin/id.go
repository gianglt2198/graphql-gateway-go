package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/gianglt2198/federation-go/package/utils"
)

type IDMixin struct {
	mixin.Schema
	Prefix string
}

// Fields defines fields for the mixin.
func (m IDMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").DefaultFunc(func() string {
			return utils.NewID(21, m.Prefix)
		}).Unique(),
	}
}

func NewID(prefix string) IDMixin {
	return IDMixin{Prefix: prefix}
}
