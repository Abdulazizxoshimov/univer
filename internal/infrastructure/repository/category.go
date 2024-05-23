package repository

import (
	"context"
	"univer/internal/entity"
)



type Category interface{
	CreateCategory(ctx context.Context, category *entity.Category) (*entity.Category, error)
	UpdateCategory(ctx context.Context, category *entity.UpdateCategory) (*entity.UpdateCategory, error)
	DeleteCategory(ctx context.Context, req *entity.DeleteReq) error
	GetCategory(ctx context.Context, params map[string]string) (*entity.Category, error)
	ListCategory(ctx context.Context, limit int, offset int) (*entity.ListCategoryRes, error)
}