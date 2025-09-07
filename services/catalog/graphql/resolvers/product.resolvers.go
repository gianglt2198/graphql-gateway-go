package resolvers

import (
	"context"
	"fmt"

	"entgo.io/contrib/entgql"

	"github.com/gianglt2198/federation-go/services/catalog/generated/ent"
	"github.com/gianglt2198/federation-go/services/catalog/generated/graph/model"
)

// FindProductEntityByID is the resolver for the findProductEntityByID field.
func (r *entityResolver) FindProductEntityByID(ctx context.Context, id string) (*model.ProductEntity, error) {
	panic(fmt.Errorf("not implemented: FindProductEntityByID - findProductEntityByID"))
}

// Products is the resolver for the products field.
func (r *queryResolver) Products(ctx context.Context, after *entgql.Cursor[string], first *int, before *entgql.Cursor[string], last *int, orderBy []*ent.ProductOrder, where *model.ProductFilter) (*model.ProductPaginatedConnection, error) {
	return r.productService.FindProducts(ctx, after, first, before, last, orderBy, where)
}

// CreateProduct is the resolver for the createProduct field.
func (r *mutationResolver) CatalogCreateProduct(ctx context.Context, input ent.CreateProductInput) (bool, error) {
	_, err := r.productService.CreateProduct(ctx, input)
	if err != nil {
		return false, err
	}
	return true, nil
}

// UpdateProduct is the resolver for the updateProduct field.
func (r *mutationResolver) CatalogUpdateProduct(ctx context.Context, id string, input ent.UpdateProductInput) (bool, error) {
	_, err := r.productService.UpdateProduct(ctx, id, input)
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteProduct is the resolver for the deleteProduct field.
func (r *mutationResolver) CatalogDeleteProduct(ctx context.Context, id string) (bool, error) {
	err := r.productService.DeleteProduct(ctx, id)
	if err != nil {
		return false, err
	}
	return true, nil
}
