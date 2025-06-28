package resolvers

import (
	"context"

	"github.com/gianglt2198/federation-go/services/account/generated/graph/model"
)

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
