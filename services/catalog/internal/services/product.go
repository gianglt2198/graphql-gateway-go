package services

import (
	"context"

	"entgo.io/contrib/entgql"
	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/utils"
	"github.com/gianglt2198/federation-go/services/catalog/generated/ent"
	"github.com/gianglt2198/federation-go/services/catalog/generated/ent/category"
	"github.com/gianglt2198/federation-go/services/catalog/generated/ent/product"
	"github.com/gianglt2198/federation-go/services/catalog/generated/graph/model"
	"github.com/gianglt2198/federation-go/services/catalog/internal/repos"
	"github.com/samber/lo"
	"go.uber.org/fx"
)

type (
	productService struct {
		log *monitoring.Logger

		productRepository repos.ProductRepository
	}

	ProductService interface {
		FindProductByID(ctx context.Context, id string) (*ent.Product, error)
		FindProducts(ctx context.Context, after *entgql.Cursor[string], first *int, before *entgql.Cursor[string], last *int, orderBy []*ent.ProductOrder, where *model.ProductFilter) (*ent.ProductConnection, error)

		CreateProduct(ctx context.Context, product ent.CreateProductInput) (*ent.Product, error)
		UpdateProduct(ctx context.Context, id string, product ent.UpdateProductInput) (*ent.Product, error)
		DeleteProduct(ctx context.Context, id string) error
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

func (s *productService) FindProductByID(ctx context.Context, id string) (*ent.Product, error) {
	return s.productRepository.FindOneWithPredicates(ctx, s.productRepository.Query(ctx), product.IDEQ(id))
}

func (s *productService) FindProducts(ctx context.Context, after *entgql.Cursor[string], first *int, before *entgql.Cursor[string], last *int, orderBy []*ent.ProductOrder, where *model.ProductFilter) (*ent.ProductConnection, error) {
	filter := func(q *ent.ProductQuery) (*ent.ProductQuery, error) {
		if where != nil {
			if len(where.Ids) > 0 {
				q = q.Where(product.IDIn(where.Ids...))
			}
			if where.CategoryIDs != nil {
				q = q.Where(product.HasCategoriesWith(category.IDIn(where.CategoryIDs...)))
			}
			if where.Name != nil {
				q = q.Where(product.NameContains(lo.FromPtr(where.Name)))
			}
		}
		return q, nil
	}

	products, err := s.productRepository.Query(ctx).Paginate(ctx, after, first, before, last, ent.WithProductOrder(orderBy), ent.WithProductFilter(filter))
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (s *productService) CreateProduct(ctx context.Context, product ent.CreateProductInput) (*ent.Product, error) {
	ctx = utils.ApplyUserIDWithContext(ctx, "system")
	return s.productRepository.CreateOne(ctx, product)
}

func (s *productService) UpdateProduct(ctx context.Context, id string, product ent.UpdateProductInput) (*ent.Product, error) {
	ctx = utils.ApplyUserIDWithContext(ctx, "system")
	return s.productRepository.UpdateOne(ctx, id, product)
}

func (s *productService) DeleteProduct(ctx context.Context, id string) error {
	ctx = utils.ApplyUserIDWithContext(ctx, "system")
	return s.productRepository.DeleteOne(ctx, id, nil)
}
