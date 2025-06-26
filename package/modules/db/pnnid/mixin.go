package pnnid

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/gianglt2198/federation-go/package/utils"
)

// MixinWithPrefix creates a Mixin that encodes the provided prefix.
func MixinWithPrefix(prefix string) *Mixin {
	return &Mixin{prefix: prefix}
}

// Mixin defines an ent Mixin that captures the PNNID prefix for a type.
type Mixin struct {
	mixin.Schema
	prefix string
}

// Fields provides the id field.
func (m Mixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(func() string { return utils.NewID(21, m.prefix) }),
	}
}

// Annotation captures the id prefix for a type.
type Annotation struct {
	Prefix string
}

// Name implements the ent Annotation interface.
func (a Annotation) Name() string {
	return "PNNID"
}

// Annotations returns the annotations for a Mixin instance.
func (m Mixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		Annotation{Prefix: m.prefix},
	}
}
