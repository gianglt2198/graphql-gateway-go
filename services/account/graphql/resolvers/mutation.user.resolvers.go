package resolvers

import (
	"context"
	"fmt"

	"github.com/gianglt2198/federation-go/package/modules/db/pnnid"
	"github.com/gianglt2198/federation-go/services/account/generated/ent"
)

// AccountCreateUser is the resolver for the accountCreateUser field.
func (r *mutationResolver) AccountCreateUser(ctx context.Context, input ent.CreateUserInput) (bool, error) {
	panic(fmt.Errorf("not implemented: AccountCreateUser - accountCreateUser"))
}

// AccountUpdateUser is the resolver for the accountUpdateUser field.
func (r *mutationResolver) AccountUpdateUser(ctx context.Context, id pnnid.ID, input ent.UpdateUserInput) (bool, error) {
	panic(fmt.Errorf("not implemented: AccountUpdateUser - accountUpdateUser"))
}

// AccountDeleteUser is the resolver for the accountDeleteUser field.
func (r *mutationResolver) AccountDeleteUser(ctx context.Context, id pnnid.ID) (bool, error) {
	panic(fmt.Errorf("not implemented: AccountDeleteUser - accountDeleteUser"))
}
