package resolvers

import (
	"context"

	"github.com/gianglt2198/federation-go/package/modules/db/pnnid"

	"github.com/gianglt2198/federation-go/services/account/generated/ent"
)

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
