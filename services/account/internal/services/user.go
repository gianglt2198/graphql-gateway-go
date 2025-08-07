package services

import (
	"context"

	"entgo.io/contrib/entgql"
	"github.com/samber/lo"
	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/utils"

	"github.com/gianglt2198/federation-go/services/account/generated/ent"
	"github.com/gianglt2198/federation-go/services/account/generated/ent/user"
	"github.com/gianglt2198/federation-go/services/account/generated/graph/model"
	"github.com/gianglt2198/federation-go/services/account/internal/repos"
)

type (
	userService struct {
		log *monitoring.Logger

		userRepository repos.UserRepository
	}

	UserService interface {
		FindUserByID(ctx context.Context, id string) (*model.UserEntity, error)
		FindUserByEmail(ctx context.Context, email string) (*model.UserEntity, error)
		FindUsers(ctx context.Context, after *entgql.Cursor[string], first *int, before *entgql.Cursor[string], last *int, orderBy []*ent.UserOrder, where *model.UserFilter) (*model.UserPaginatedConnection, error)

		CreateUser(ctx context.Context, input ent.CreateUserInput) (*ent.User, error)
		UpdateUser(ctx context.Context, id string, input ent.UpdateUserInput) (*ent.User, error)
		DeleteUser(ctx context.Context, id string) error
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

func (s *userService) FindUserByID(ctx context.Context, id string) (*model.UserEntity, error) {
	user, err := s.userRepository.FindOneWithPredicates(ctx, s.userRepository.WithCollectFields(ctx), user.IDEQ(id))
	if err != nil {
		return nil, err
	}

	userEntity, err := utils.ConvertTo[ent.User, model.UserEntity](user)
	if err != nil {
		return nil, err
	}

	return userEntity, nil
}

func (s *userService) FindUserByEmail(ctx context.Context, email string) (*model.UserEntity, error) {
	user, err := s.userRepository.FindOneWithPredicates(ctx, s.userRepository.Query(ctx), user.EmailEQ(email))
	if err != nil {
		return nil, err
	}

	userEntity, err := utils.StructToStruct[ent.User, model.UserEntity](user)
	if err != nil {
		return nil, err
	}

	return userEntity, nil
}

func (s *userService) FindUsers(ctx context.Context, after *entgql.Cursor[string], first *int, before *entgql.Cursor[string], last *int, orderBy []*ent.UserOrder, where *model.UserFilter) (*model.UserPaginatedConnection, error) {
	filter := func(q *ent.UserQuery) (*ent.UserQuery, error) {
		if where != nil {
			if len(where.Ids) > 0 {
				q = q.Where(user.IDIn(where.Ids...))
			}
			if where.Email != nil {
				q = q.Where(user.EmailEQ(lo.FromPtr(where.Email)))
			}
			if where.Username != nil {
				q = q.Where(user.UsernameEQ(lo.FromPtr(where.Username)))
			}
		}
		return q, nil
	}

	users, err := s.userRepository.Query(ctx).Paginate(ctx, after, first, before, last, ent.WithUserFilter(filter), ent.WithUserOrder(orderBy))
	if err != nil {
		return nil, err
	}

	users.PageInfo = entgql.PageInfo[string]{
		HasNextPage:     users.PageInfo.HasNextPage,
		HasPreviousPage: users.PageInfo.HasPreviousPage,
		StartCursor:     users.PageInfo.StartCursor,
		EndCursor:       users.PageInfo.EndCursor,
	}

	list := make([]*model.UserPaginatedEdge, len(users.Edges))
	for i, edge := range users.Edges {
		user := edge.Node

		userEntity, err := utils.ConvertTo[ent.User, model.UserEntity](user)
		if err != nil {
			return nil, err
		}

		list[i] = &model.UserPaginatedEdge{
			Cursor: edge.Cursor,
			Node:   userEntity,
		}
	}

	return &model.UserPaginatedConnection{
		Edges:      list,
		PageInfo:   lo.ToPtr(users.PageInfo),
		TotalCount: users.TotalCount,
	}, nil
}

func (s *userService) CreateUser(ctx context.Context, input ent.CreateUserInput) (*ent.User, error) {
	return s.userRepository.CreateOne(ctx, input)
}

func (s *userService) UpdateUser(ctx context.Context, id string, input ent.UpdateUserInput) (*ent.User, error) {
	return s.userRepository.UpdateOne(ctx, id, input)
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	return s.userRepository.DeleteOne(ctx, id, nil)
}
