package repos

import (
	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"

	"github.com/gianglt2198/federation-go/services/catalog/generated/ent"
)

type (
	categoryRepository struct {
		ent.CategoryRepository

		log *logging.Logger
		db  *ent.Client
	}

	CategoryRepository interface {
		ent.CategoryRepository
	}
)

type CategoryRepositoryParams struct {
	fx.In

	Log *logging.Logger
	Db  *ent.Client
}

type CategoryRepositoryResult struct {
	fx.Out

	CategoryRepository CategoryRepository
}

func NewCategoryRepository(params CategoryRepositoryParams) CategoryRepositoryResult {
	return CategoryRepositoryResult{
		CategoryRepository: &categoryRepository{
			CategoryRepository: ent.NewCategoryRepository(params.Log, params.Db),
			log:                params.Log,
			db:                 params.Db,
		},
	}
}
