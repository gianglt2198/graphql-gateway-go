package repos

import (
	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/services/account/generated/ent"

	"go.uber.org/fx"
)

type (
	userRepository struct {
		ent.UserRepository

		log *monitoring.Logger
		db  *ent.Client
	}

	UserRepository interface {
		ent.UserRepository
	}
)

type UserRepositoryParams struct {
	fx.In

	Log *monitoring.Logger
	Db  *ent.Client
}

type UserRepositoryResult struct {
	fx.Out

	UserRepository UserRepository
}

func NewUserRepository(params UserRepositoryParams) UserRepositoryResult {
	return UserRepositoryResult{
		UserRepository: &userRepository{
			UserRepository: ent.NewUserRepository(params.Log, params.Db),
			log:            params.Log,
			db:             params.Db,
		},
	}
}
