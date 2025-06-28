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
	categoryService struct {
		log *monitoring.Logger

		categoryRepository repos.CategoryRepository
	}

	CategoryService interface {
		FindCategoryByID(ctx context.Context, id string) (*ent.Category, error)
		FindCategories(ctx context.Context, after *entgql.Cursor[string], first *int, before *entgql.Cursor[string], last *int, orderBy []*ent.CategoryOrder, where *model.CategoryFilter) (*ent.CategoryConnection, error)

		CreateCategory(ctx context.Context, product ent.CreateCategoryInput) (*ent.Category, error)
		UpdateCategory(ctx context.Context, id string, product ent.UpdateCategoryInput) (*ent.Category, error)
		DeleteCategory(ctx context.Context, id string) error
	}
)

type CategoryServiceParams struct {
	fx.In

	Log *monitoring.Logger

	CategoryRepository repos.CategoryRepository
}

type CategoryServiceResult struct {
	fx.Out

	CategoryService CategoryService
}

func NewCategoryService(params CategoryServiceParams) CategoryServiceResult {
	return CategoryServiceResult{
		CategoryService: &categoryService{
			log:                params.Log,
			categoryRepository: params.CategoryRepository,
		},
	}
}

func (s *categoryService) FindCategoryByID(ctx context.Context, id string) (*ent.Category, error) {
	return s.categoryRepository.FindOneWithPredicates(ctx, s.categoryRepository.Query(ctx), category.IDEQ(id))
}

func (s *categoryService) FindCategories(ctx context.Context, after *entgql.Cursor[string], first *int, before *entgql.Cursor[string], last *int, orderBy []*ent.CategoryOrder, where *model.CategoryFilter) (*ent.CategoryConnection, error) {
	filter := func(q *ent.CategoryQuery) (*ent.CategoryQuery, error) {
		if where != nil {
			if len(where.Ids) > 0 {
				q = q.Where(category.IDIn(where.Ids...))
			}
			if where.Name != nil {
				q = q.Where(category.NameContains(lo.FromPtr(where.Name)))
			}
			if where.ProductIDs != nil {
				q = q.Where(category.HasProductsWith(product.IDIn(where.ProductIDs...)))
			}
		}
		return q, nil
	}

	categories, err := s.categoryRepository.Query(ctx).Paginate(ctx, after, first, before, last, ent.WithCategoryOrder(orderBy), ent.WithCategoryFilter(filter))
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *categoryService) CreateCategory(ctx context.Context, category ent.CreateCategoryInput) (*ent.Category, error) {
	ctx = utils.ApplyUserIDWithContext(ctx, "system")
	return s.categoryRepository.CreateOne(ctx, category)
}

func (s *categoryService) UpdateCategory(ctx context.Context, id string, category ent.UpdateCategoryInput) (*ent.Category, error) {
	ctx = utils.ApplyUserIDWithContext(ctx, "system")
	return s.categoryRepository.UpdateOne(ctx, id, category)
}

func (s *categoryService) DeleteCategory(ctx context.Context, id string) error {
	ctx = utils.ApplyUserIDWithContext(ctx, "system")
	return s.categoryRepository.DeleteOne(ctx, id, nil)
}
