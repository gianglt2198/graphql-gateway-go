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
	"github.com/gianglt2198/federation-go/services/catalog/internal/dtos"
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
		FindCategoryByID(ctx context.Context, id string) (*model.CategoryEntity, error)
		FindCategories(ctx context.Context, after *entgql.Cursor[string], first *int, before *entgql.Cursor[string], last *int, orderBy []*ent.CategoryOrder, where *model.CategoryFilter) (*model.CategoryPaginatedConnection, error)

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

func (s *categoryService) FindCategoryByID(ctx context.Context, id string) (*model.CategoryEntity, error) {
	category, err := s.categoryRepository.FindOneWithPredicates(ctx, s.categoryRepository.WithCollectFields(ctx), category.IDEQ(id))
	if err != nil {
		return nil, err
	}

	return dtos.ToCategoryEntity(category)
}

func (s *categoryService) FindCategories(ctx context.Context, after *entgql.Cursor[string], first *int, before *entgql.Cursor[string], last *int, orderBy []*ent.CategoryOrder, where *model.CategoryFilter) (*model.CategoryPaginatedConnection, error) {
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

	categories.PageInfo = entgql.PageInfo[string]{
		HasNextPage:     categories.PageInfo.HasNextPage,
		HasPreviousPage: categories.PageInfo.HasPreviousPage,
		StartCursor:     categories.PageInfo.StartCursor,
		EndCursor:       categories.PageInfo.EndCursor,
	}

	list := make([]*model.CategoryPaginatedEdge, len(categories.Edges))
	for i, edge := range categories.Edges {
		category := edge.Node

		categoryEntity, err := dtos.ToCategoryEntity(category)
		if err != nil {
			return nil, err
		}

		list[i] = &model.CategoryPaginatedEdge{
			Cursor: edge.Cursor,
			Node:   categoryEntity,
		}
	}

	return &model.CategoryPaginatedConnection{
		Edges:      list,
		PageInfo:   lo.ToPtr(categories.PageInfo),
		TotalCount: categories.TotalCount,
	}, nil
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
