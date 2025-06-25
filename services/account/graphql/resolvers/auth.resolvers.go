package resolvers

import (
	"context"
	"fmt"

	"entgo.io/contrib/entgql"
	"github.com/gianglt2198/federation-go/package/modules/db/pnnid"

	"github.com/gianglt2198/federation-go/services/account/generated/ent"
	"github.com/gianglt2198/federation-go/services/account/generated/graph/model"
)

// Sessions is the resolver for the sessions field.
func (r *queryResolver) Sessions(ctx context.Context, after *entgql.Cursor[pnnid.ID], first *int, before *entgql.Cursor[pnnid.ID], last *int, where *ent.SessionWhereInput) (*ent.SessionConnection, error) {
	return r.db.Session.Query().Paginate(ctx, after, first, before, last, ent.WithSessionFilter(where.Filter))
}

// AccountAuthRegister is the resolver for the accountAuthRegister field.
func (r *mutationResolver) AccountAuthRegister(ctx context.Context, input model.RegisterInput) (bool, error) {
	_, err := r.authService.Register(ctx, input)
	if err != nil {
		return false, err
	}
	return true, nil
}

// AccountAuthLogin is the resolver for the accountAuthLogin field.
func (r *mutationResolver) AccountAuthLogin(ctx context.Context, input model.LoginInput) (*model.LoginEntity, error) {
	token, err := r.authService.Login(ctx, input)
	if err != nil {
		return nil, err
	}

	return &model.LoginEntity{Token: token}, nil
}

// AccountAuthLogout is the resolver for the accountAuthLogout field.
func (r *mutationResolver) AccountAuthLogout(ctx context.Context) (bool, error) {
	err := r.authService.Logout(ctx, pnnid.ID(ctx.Value("session_id").(string)))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) AccountAuthVerify(ctx context.Context) (*model.AuthVerifyEntity, error) {
	panic(fmt.Errorf("not implemented: AccountAuthVerify - accountAuthVerify"))
}
