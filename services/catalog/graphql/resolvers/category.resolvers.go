package resolvers

import (
	"context"
	"fmt"

	"entgo.io/contrib/entgql"
	"github.com/gianglt2198/federation-go/services/catalog/generated/ent"
	"github.com/gianglt2198/federation-go/services/catalog/generated/graph/model"
)

// FindCategoryEntityByID is the resolver for the findCategoryEntityByID field.
func (r *entityResolver) FindCategoryEntityByID(ctx context.Context, id string) (*model.CategoryEntity, error) {
	panic(fmt.Errorf("not implemented: FindCategoryEntityByID - findCategoryEntityByID"))
}

// Category is the resolver for the category field.
func (r *queryResolver) Category(ctx context.Context, id string) (*model.CategoryEntity, error) {
	return r.categoryService.FindCategoryByID(ctx, id)
}

// Categories is the resolver for the categories field.
func (r *queryResolver) Categories(ctx context.Context, after *entgql.Cursor[string], first *int, before *entgql.Cursor[string], last *int, orderBy []*ent.CategoryOrder, where *model.CategoryFilter) (*model.CategoryPaginatedConnection, error) {
	return r.categoryService.FindCategories(ctx, after, first, before, last, orderBy, where)
}

// CreateCategory is the resolver for the createCategory field.
func (r *mutationResolver) CatalogCreateCategory(ctx context.Context, input ent.CreateCategoryInput) (bool, error) {
	_, err := r.categoryService.CreateCategory(ctx, input)
	if err != nil {
		return false, err
	}
	return true, nil
}

// UpdateCategory is the resolver for the updateCategory field.
func (r *mutationResolver) CatalogUpdateCategory(ctx context.Context, id string, input ent.UpdateCategoryInput) (bool, error) {
	_, err := r.categoryService.UpdateCategory(ctx, id, input)
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteCategory is the resolver for the deleteCategory field.
func (r *mutationResolver) CatalogDeleteCategory(ctx context.Context, id string) (bool, error) {
	err := r.categoryService.DeleteCategory(ctx, id)
	if err != nil {
		return false, err
	}
	return true, nil
}
