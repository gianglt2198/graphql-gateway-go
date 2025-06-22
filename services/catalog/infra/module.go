package infra

import (
	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/services/catalog/graphql"
	"github.com/gianglt2198/federation-go/services/catalog/internal/repos"
	"github.com/gianglt2198/federation-go/services/catalog/internal/services"
)

var Module = fx.Module("infra",
	dbModule,
	repos.Module,
	services.Module,
	graphql.Module,
)
