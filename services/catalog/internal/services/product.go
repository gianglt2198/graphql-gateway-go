package services

import (
	"context"

	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/modules/db/pnnid"
	"github.com/gianglt2198/federation-go/services/catalog/generated/ent"
	"github.com/gianglt2198/federation-go/services/catalog/internal/repos"
	"go.uber.org/fx"
)

type (
	productService struct {
		log *monitoring.Logger

		productRepository repos.ProductRepository
	}

	ProductService interface {
		CreateProduct(ctx context.Context, product *ent.CreateProductInput) (*ent.Product, error)
		UpdateProduct(ctx context.Context, id pnnid.ID, product *ent.UpdateProductInput) (*ent.Product, error)
		DeleteProduct(ctx context.Context, id pnnid.ID) error
	}
)

type ProductServiceParams struct {
	fx.In

	Log *monitoring.Logger

	ProductRepository repos.ProductRepository
}

type ProductServiceResult struct {
	fx.Out

	ProductService ProductService
}

func NewProductService(params ProductServiceParams) ProductServiceResult {
	return ProductServiceResult{
		ProductService: &productService{
			log:               params.Log,
			productRepository: params.ProductRepository,
		},
	}
}

func (s *productService) CreateProduct(ctx context.Context, product *ent.CreateProductInput) (*ent.Product, error) {
	return s.productRepository.CreateOne(ctx, *product)
}

func (s *productService) UpdateProduct(ctx context.Context, id pnnid.ID, product *ent.UpdateProductInput) (*ent.Product, error) {
	return s.productRepository.UpdateOne(ctx, id, *product)
}

func (s *productService) DeleteProduct(ctx context.Context, id pnnid.ID) error {
	return s.productRepository.DeleteOne(ctx, id, nil)
}
