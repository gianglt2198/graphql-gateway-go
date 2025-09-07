package dtos

import (
	"github.com/samber/lo"

	"github.com/gianglt2198/federation-go/package/utils"

	"github.com/gianglt2198/federation-go/services/catalog/generated/ent"
	"github.com/gianglt2198/federation-go/services/catalog/generated/graph/model"
)

func ToCategoryEntity(category *ent.Category) (*model.CategoryEntity, error) {
	if category == nil {
		return nil, nil
	}
	categoryEntity, err := utils.ConvertTo[ent.Category, model.CategoryEntity](category)
	if err != nil {
		return nil, err
	}
	if categoryEntity == nil {
		return nil, nil
	}

	if categoryEntity.CreatedBy != nil {
		categoryEntity.UserCreatedBy = &model.UserEntity{
			ID: lo.FromPtr(categoryEntity.CreatedBy),
		}
	}

	if categoryEntity.UpdatedBy != nil {
		categoryEntity.UserUpdatedBy = &model.UserEntity{
			ID: lo.FromPtr(categoryEntity.UpdatedBy),
		}
	}

	return categoryEntity, nil
}
