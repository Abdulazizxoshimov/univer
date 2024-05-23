package usecase

import (
	"context"
	"time"
	"univer/internal/entity"
	"univer/internal/infrastructure/repository"
	"univer/internal/pkg/otlp"
)

const (
	serviceNameCategoryService = "categoryServiceRepo"
	spanNameCategoryService    = "categorySpanRepo"
)

type Category interface{
	CreateCategory(ctx context.Context, category *entity.Category) (*entity.Category, error)
	UpdateCategory(ctx context.Context, category *entity.UpdateCategory) (*entity.UpdateCategory, error)
	DeleteCategory(ctx context.Context, req *entity.DeleteReq) error
	GetCategory(ctx context.Context, req *entity.GetReq) (*entity.Category, error)
	ListCategory(ctx context.Context, req *entity.ListReq) (*entity.ListCategoryRes, error)
}

type categoryService struct{
	BaseUseCase
	ctxTimeout time.Duration
	repo    repository.Category
}


func NewCategoryService (ctxTimeout time.Duration, repo repository.Category)Category{
	return categoryService{
		ctxTimeout: ctxTimeout,
		repo: repo,
	}

}

func (p categoryService) CreateCategory(ctx context.Context, Category *entity.Category) (*entity.Category, error) {
	ctx, span := otlp.Start(ctx, serviceNameCategoryService, spanNameCategoryService + "CreateCategory")
	defer span.End()

	p.beforeRequest(&Category.Id, &Category.CreatedAt, &Category.UpdatedAt, nil)

	return p.repo.CreateCategory(ctx, Category)
}
func (p categoryService) UpdateCategory(ctx context.Context, category *entity.UpdateCategory) (*entity.UpdateCategory, error) {
	ctx, span := otlp.Start(ctx, serviceNameCategoryService, spanNameCategoryService + "UpdateCategory")
	defer span.End()

	p.beforeRequest(nil, nil, &category.UpdatedAt, nil)

	return p.repo.UpdateCategory(ctx, category)
}
func (p categoryService) DeleteCategory(ctx context.Context, req *entity.DeleteReq) error {
	ctx, span := otlp.Start(ctx,  serviceNameCategoryService, spanNameCategoryService + "DeleteCategory")
	defer span.End()

	p.beforeRequest(nil, nil, nil, &req.DeletedAt)

	return p.repo.DeleteCategory(ctx, req)
}
func (p categoryService) GetCategory(ctx context.Context, req *entity.GetReq) (*entity.Category, error) {
	ctx, span := otlp.Start(ctx, serviceNameCategoryService, spanNameCategoryService + "GetCategory")
	defer span.End()

	return p.repo.GetCategory(ctx, req.Filter)
}
func (p categoryService) ListCategory(ctx context.Context, req *entity.ListReq) (*entity.ListCategoryRes, error) {
	ctx, span := otlp.Start(ctx, serviceNameCategoryService, spanNameCategoryService + "ListCategory")
	defer span.End()

	return p.repo.ListCategory(ctx, req.Limit, req.Offset)
}
