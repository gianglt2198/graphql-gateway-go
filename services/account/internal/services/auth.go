package services

import (
	"context"

	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/modules/db/pnnid"
	"github.com/gianglt2198/federation-go/services/account/internal/repos"
	"go.uber.org/fx"
)

type (
	authService struct {
		log *monitoring.Logger

		sessionRepository repos.SessionRepository
	}

	AuthService interface {
		// Login(ctx context.Context, input *model.LoginInput) (*ent.Session, error)
		Logout(ctx context.Context, sessionID pnnid.ID) error
	}
)

type AuthServiceParams struct {
	fx.In

	Log *monitoring.Logger

	SessionRepository repos.SessionRepository
}

type AuthServiceResult struct {
	fx.Out

	AuthService AuthService
}

func NewAuthService(params AuthServiceParams) AuthServiceResult {
	return AuthServiceResult{
		AuthService: &authService{
			log:               params.Log,
			sessionRepository: params.SessionRepository,
		},
	}
}

// func (s *authService) Login(ctx context.Context, input *model.LoginInput) (*ent.Session, error) {
// 	return s.sessionRepository.CreateOne(ctx, *input)
// }

func (s *authService) Logout(ctx context.Context, sessionID pnnid.ID) error {
	return s.sessionRepository.DeleteOne(ctx, sessionID, nil)
}
