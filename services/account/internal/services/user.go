package services

import (
	"context"

	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/modules/db/pnnid"

	"github.com/gianglt2198/federation-go/services/account/generated/ent"
	"github.com/gianglt2198/federation-go/services/account/internal/repos"
)

type (
	userService struct {
		log *monitoring.Logger

		userRepository repos.UserRepository
	}

	UserService interface {
		CreateUser(ctx context.Context, input ent.CreateUserInput) (*ent.User, error)
		UpdateUser(ctx context.Context, id pnnid.ID, input ent.UpdateUserInput) (*ent.User, error)
		DeleteUser(ctx context.Context, id pnnid.ID) error
	}
)

type UserServiceParams struct {
	fx.In

	Log *monitoring.Logger

	UserRepository repos.UserRepository
}

type UserServiceResult struct {
	fx.Out

	UserService UserService
}

func NewUserService(params UserServiceParams) UserServiceResult {
	return UserServiceResult{
		UserService: &userService{
			log:            params.Log,
			userRepository: params.UserRepository,
		},
	}
}

func (s *userService) CreateUser(ctx context.Context, input ent.CreateUserInput) (*ent.User, error) {
	return s.userRepository.CreateOne(ctx, input)
}

func (s *userService) UpdateUser(ctx context.Context, id pnnid.ID, input ent.UpdateUserInput) (*ent.User, error) {
	return s.userRepository.UpdateOne(ctx, id, input)
}

func (s *userService) DeleteUser(ctx context.Context, id pnnid.ID) error {
	return s.userRepository.DeleteOne(ctx, id, nil)
}
