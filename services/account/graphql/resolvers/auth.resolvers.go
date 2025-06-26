package resolvers

import (
	"context"

	"entgo.io/contrib/entgql"

	"github.com/gianglt2198/federation-go/services/account/generated/ent"
	"github.com/gianglt2198/federation-go/services/account/generated/graph/model"
)

// Sessions is the resolver for the sessions field.
func (r *queryResolver) Sessions(ctx context.Context, after *entgql.Cursor[string], first *int, before *entgql.Cursor[string], last *int, where *ent.SessionWhereInput) (*ent.SessionConnection, error) {
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
	err := r.authService.Logout(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) AccountAuthVerify(ctx context.Context, input model.AuthVerifyInput) (*model.AuthVerifyEntity, error) {
	return r.authService.AuthVerify(ctx, input)
}
