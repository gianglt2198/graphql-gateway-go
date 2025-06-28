package resolvers

import (
	"context"

	"entgo.io/contrib/entgql"

	"github.com/gianglt2198/federation-go/services/account/generated/ent"
	"github.com/gianglt2198/federation-go/services/account/generated/graph/model"
)

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context, after *entgql.Cursor[string], first *int, before *entgql.Cursor[string], last *int, orderBy []*ent.UserOrder, where *model.UserFilter) (*ent.UserConnection, error) {
	return r.userService.FindUsers(ctx, after, first, before, last, orderBy, where)
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
func (r *mutationResolver) AccountUpdateUser(ctx context.Context, id string, input ent.UpdateUserInput) (bool, error) {
	_, err := r.userService.UpdateUser(ctx, id, input)
	if err != nil {
		return false, err
	}
	return true, nil
}

// AccountDeleteUser is the resolver for the accountDeleteUser field.
func (r *mutationResolver) AccountDeleteUser(ctx context.Context, id string) (bool, error) {
	err := r.userService.DeleteUser(ctx, id)
	if err != nil {
		return false, err
	}
	return true, nil
}
