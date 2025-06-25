package resolvers

import (
	"context"

	"entgo.io/contrib/entgql"
	"github.com/gianglt2198/federation-go/package/modules/db/pnnid"

	"github.com/gianglt2198/federation-go/services/account/generated/ent"
)

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context, after *entgql.Cursor[pnnid.ID], first *int, before *entgql.Cursor[pnnid.ID], last *int, orderBy []*ent.UserOrder, where *ent.UserWhereInput) (*ent.UserConnection, error) {
	return r.db.User.Query().Paginate(ctx, after, first, before, last, ent.WithUserFilter(where.Filter))
}

// AccountCreateUser is the resolver for the accountCreateUser field.
func (r *mutationResolver) AccountCreateUser(ctx context.Context, input ent.CreateUserInput) (bool, error) {
	_, err := r.userService.CreateUser(ctx, input)
	if err != nil {
		return false, err
	}
	return true, nil
}

// AccountUpdateUser is the resolver for the accountUpdateUser field.
func (r *mutationResolver) AccountUpdateUser(ctx context.Context, id pnnid.ID, input ent.UpdateUserInput) (bool, error) {
	_, err := r.userService.UpdateUser(ctx, id, input)
	if err != nil {
		return false, err
	}
	return true, nil
}

// AccountDeleteUser is the resolver for the accountDeleteUser field.
func (r *mutationResolver) AccountDeleteUser(ctx context.Context, id pnnid.ID) (bool, error) {
	err := r.userService.DeleteUser(ctx, id)
	if err != nil {
		return false, err
	}
	return true, nil
}
