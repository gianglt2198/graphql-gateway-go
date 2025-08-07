package dtos

import (
	"github.com/gianglt2198/federation-go/package/utils"

	"github.com/gianglt2198/federation-go/services/catalog/generated/ent"
	"github.com/gianglt2198/federation-go/services/catalog/generated/graph/model"
)

func ToProductEntity(product *ent.Product) (*model.ProductEntity, error) {
	if product == nil {
		return nil, nil
	}
	productEntity, err := utils.ConvertTo[ent.Product, model.ProductEntity](product)
	if err != nil {
		return nil, err
	}
	if productEntity == nil {
		return nil, nil
	}

	return productEntity, nil
}
