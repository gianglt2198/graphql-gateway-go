package mixin

import (
	"context"
	"fmt"
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/gianglt2198/federation-go/package/utils"
	"github.com/gianglt2198/federation-go/package/utils/reflection"
)

// -------------------------------------------------
// Mixin definition

// AuthorMixin implements the ent.Mixin for sharing
// time fields with package schemas.
type AuthorMixin struct {
	// We embed the `mixin.Schema` to avoid
	// implementing the rest of the methods.
	mixin.Schema
}

type (
	SetUpdatedByKey struct{}
	SetCreatedByKey struct{}
)

const CreateByColumnName = "created_by"
const UpdatedByColumnName = "updated_by"

func (AuthorMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Annotations(
				entgql.OrderField("createdAt"),
			),
		field.String(CreateByColumnName).
			Immutable().
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("createdBy"),
			),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Annotations(
				entgql.OrderField("updatedAt"),
			),
		field.String(UpdatedByColumnName).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("updatedBy"),
			),
	}
}

// SkipSetUpdatedBy returns a new context that skips the soft-delete interceptor/mutators.
func SkipSetUpdatedBy(parent context.Context) context.Context {
	return context.WithValue(parent, SetUpdatedByKey{}, true)
}

// SkipSetCreatedBy returns a new context that skips the soft-delete interceptor/mutators.
func SkipSetCreatedBy(parent context.Context) context.Context {
	return context.WithValue(parent, SetCreatedByKey{}, true)
}

func (AuthorMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				if m.Op().Is(ent.OpCreate) {
					if skip, _ := ctx.Value(SetCreatedByKey{}).(bool); skip {
						return next.Mutate(ctx, m)
					}
					if _, ok := m.(interface {
						SetCreatedBy(string)
					}); ok {
						userID := utils.GetUserIDFromCtx(ctx)
						if userID == "" {
							return nil, fmt.Errorf("user ID not found in context")
						}
						reflection.CallMethod("SetCreatedBy", m, userID)
					}
				}

				// Process set UpdatedBy.
				if m.Op().Is(ent.OpUpdate) || m.Op().Is(ent.OpUpdateOne) {
					if skip, _ := ctx.Value(SetUpdatedByKey{}).(bool); skip {
						return next.Mutate(ctx, m)
					}
					if _, ok := m.(interface {
						SetUpdatedBy(string)
					}); ok {
						userID := utils.GetUserIDFromCtx(ctx)
						if userID == "" {
							return nil, fmt.Errorf("user ID not found in context")
						}
						reflection.CallMethod("SetUpdatedBy", m, userID)
					}
				}

				return next.Mutate(ctx, m)
			})
		},
	}
}
