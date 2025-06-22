package resolvers

import (
	"context"
	"fmt"

	"github.com/gianglt2198/federation-go/services/account/generated/graph/model"
)

// AccountAuthLogin is the resolver for the accountAuthLogin field.
func (r *mutationResolver) AccountAuthLogin(ctx context.Context, input model.LoginInput) (string, error) {
	panic(fmt.Errorf("not implemented: AccountAuthLogin - accountAuthLogin"))
}

// AccountAuthLogout is the resolver for the accountAuthLogout field.
func (r *mutationResolver) AccountAuthLogout(ctx context.Context) (bool, error) {
	panic(fmt.Errorf("not implemented: AccountAuthLogout - accountAuthLogout"))
}
