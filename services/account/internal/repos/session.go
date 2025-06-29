package repos

import (
	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/package/infras/monitoring"

	"github.com/gianglt2198/federation-go/services/account/generated/ent"
)

type (
	sessionRepository struct {
		ent.SessionRepository

		log *monitoring.Logger
		db  *ent.Client
	}

	SessionRepository interface {
		ent.SessionRepository
	}
)

type SessionRepositoryParams struct {
	fx.In

	Log *monitoring.Logger
	Db  *ent.Client
}

type SessionRepositoryResult struct {
	fx.Out

	SessionRepository SessionRepository
}

func NewSessionRepository(params SessionRepositoryParams) SessionRepositoryResult {
	return SessionRepositoryResult{
		SessionRepository: &sessionRepository{
			SessionRepository: ent.NewSessionRepository(params.Log, params.Db),
			log:               params.Log,
			db:                params.Db,
		},
	}
}
