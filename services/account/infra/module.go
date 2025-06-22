package infra

import (
	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/services/account/graphql"
	"github.com/gianglt2198/federation-go/services/account/internal/repos"
	"github.com/gianglt2198/federation-go/services/account/internal/services"
)

var Module = fx.Module("infra",
	dbModule,
	repos.Module,
	services.Module,
	graphql.Module,
)
