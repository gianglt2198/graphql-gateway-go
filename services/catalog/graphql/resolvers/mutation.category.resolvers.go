package resolvers

import (
	"context"
	"fmt"

	"github.com/gianglt2198/federation-go/package/modules/db/pnnid"
	"github.com/gianglt2198/federation-go/services/catalog/generated/ent"
)

// CreateCategory is the resolver for the createCategory field.
func (r *mutationResolver) CreateCategory(ctx context.Context, input ent.CreateCategoryInput) (bool, error) {
	panic(fmt.Errorf("not implemented: CreateCategory - createCategory"))
}

// UpdateCategory is the resolver for the updateCategory field.
func (r *mutationResolver) UpdateCategory(ctx context.Context, id pnnid.ID, input ent.UpdateCategoryInput) (bool, error) {
	panic(fmt.Errorf("not implemented: UpdateCategory - updateCategory"))
}

// DeleteCategory is the resolver for the deleteCategory field.
func (r *mutationResolver) DeleteCategory(ctx context.Context, id pnnid.ID) (bool, error) {
	panic(fmt.Errorf("not implemented: DeleteCategory - deleteCategory"))
}
