package resolvers

import (
	"context"
	"fmt"

	"github.com/gianglt2198/federation-go/package/modules/db/pnnid"
	"github.com/gianglt2198/federation-go/services/catalog/generated/ent"
)

// CreateProduct is the resolver for the createProduct field.
func (r *mutationResolver) CreateProduct(ctx context.Context, input ent.CreateProductInput) (bool, error) {
	panic(fmt.Errorf("not implemented: CreateProduct - createProduct"))
}

// UpdateProduct is the resolver for the updateProduct field.
func (r *mutationResolver) UpdateProduct(ctx context.Context, id pnnid.ID, input ent.UpdateProductInput) (bool, error) {
	panic(fmt.Errorf("not implemented: UpdateProduct - updateProduct"))
}

// DeleteProduct is the resolver for the deleteProduct field.
func (r *mutationResolver) DeleteProduct(ctx context.Context, id pnnid.ID) (bool, error) {
	panic(fmt.Errorf("not implemented: DeleteProduct - deleteProduct"))
}
