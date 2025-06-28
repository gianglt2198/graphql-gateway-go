package repos

import (
	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/services/catalog/generated/ent"
	"go.uber.org/fx"
)

type (
	categoryRepository struct {
		ent.CategoryRepository

		log *monitoring.Logger
		db  *ent.Client
	}

	CategoryRepository interface {
		ent.CategoryRepository
	}
)

type CategoryRepositoryParams struct {
	fx.In

	Log *monitoring.Logger
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
