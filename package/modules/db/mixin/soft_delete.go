package mixin

import (
	"context"
	"fmt"
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/gianglt2198/federation-go/package/utils"
	"github.com/gianglt2198/federation-go/package/utils/reflection"
)

// -------------------------------------------------
// Mixin definition

// TimeMixin implements the ent.Mixin for sharing
// soft delete time fields with package schemas.
type SoftDeleteMixin struct {
	// We embed the `mixin.Schema` to avoid
	// implementing the rest of the methods.
	mixin.Schema
}

type (
	SoftDeleteKey    struct{}
	InterceptorQuery interface {
		WhereP(...func(*sql.Selector))
	}
)

const SoftDeletedAtColumnName = "deleted_at"

const SoftDeletedByColumnName = "deleted_by"

func (SoftDeleteMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time(SoftDeletedAtColumnName).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("deletedAt"),
			),
		field.String(SoftDeletedByColumnName).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("deletedBy"),
			),
	}
}

// SkipSoftDelete returns a new context that skips the soft-delete interceptor/mutators.
func SkipSoftDelete(parent context.Context) context.Context {
	return context.WithValue(parent, SoftDeleteKey{}, true)
}

func (d SoftDeleteMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				// Process only for delete operations.
				if m.Op().Is(ent.OpDeleteOne) || m.Op().Is(ent.OpDelete) {
					if skip, _ := ctx.Value(SoftDeleteKey{}).(bool); skip {
						return next.Mutate(ctx, m)
					}
					if _, ok := m.(interface {
						SetDeletedBy(string)
						SetDeletedAt(time.Time)
					}); ok {
						userID := utils.GetUserIDFromCtx(ctx)
						if userID == "" {
							return nil, fmt.Errorf("user ID not found in context")
						}
						reflection.CallMethod("SetOp", m, ent.OpUpdate)
						reflection.CallMethod("SetDeletedAt", m, time.Now())
						reflection.CallMethod("SetDeletedBy", m, userID)
						return next.Mutate(ctx, m)
					}
				}
				return next.Mutate(ctx, m)
			})
		},
	}
}

// Interceptors of the SoftDeleteMixin.
func (d SoftDeleteMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		ent.TraverseFunc(func(ctx context.Context, q ent.Query) error {
			// Skip soft-delete, means include soft-deleted entities.
			if skip, _ := ctx.Value(SoftDeleteKey{}).(bool); skip {
				return nil
			}

			reflection.CallMethod("Where", q, sql.FieldIsNull(SoftDeletedAtColumnName))
			return nil
		}),
	}

}
