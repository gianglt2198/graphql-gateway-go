package repos

import (
	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"

	"github.com/gianglt2198/federation-go/services/catalog/generated/ent"
)

type (
	productRepository struct {
		ent.ProductRepository

		log *logging.Logger
		db  *ent.Client
	}

	ProductRepository interface {
		ent.ProductRepository
	}
)

type ProductRepositoryParams struct {
	fx.In

	Log *logging.Logger
	Db  *ent.Client
}

type ProductRepositoryResult struct {
	fx.Out

	ProductRepository ProductRepository
}

func NewProductRepository(params ProductRepositoryParams) ProductRepositoryResult {
	return ProductRepositoryResult{
		ProductRepository: &productRepository{
			ProductRepository: ent.NewProductRepository(params.Log, params.Db),
			log:               params.Log,
			db:                params.Db,
		},
	}
}
