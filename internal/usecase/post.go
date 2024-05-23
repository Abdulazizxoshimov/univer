package usecase

import (
	"context"
	"time"
	"univer/internal/entity"
	"univer/internal/infrastructure/repository"
	"univer/internal/pkg/otlp"
)

const (
	postServiceTableName    = "posts"
	serviceNamePostsService = "postServiceUsecase"
	spanNamePostsService    = "postSpanUsecase"
)

type Post interface {
	CreatePost(ctx context.Context, post *entity.Post) (*entity.Post, error)
	UpdatePost(ctx context.Context, post *entity.PostUpdateReq) (*entity.PostUpdateReq, error)
	DeletePost(ctx context.Context, req *entity.DeleteReq) error
	GetPost(ctx context.Context, req *entity.GetReq) (*entity.Post, error)
	ListPost(ctx context.Context, req *entity.ListReq) (*entity.PostListRes, error)
	Search(ctx context.Context, req *entity.ListReq)(*entity.PostListRes, error)
}

type postService struct{
	BaseUseCase
	ctxTimeout time.Duration
	repo repository.Post
}

func NewPostService(ctxTimout time.Duration, repo repository.Post) Post{
	return  postService{
		ctxTimeout: ctxTimout,
		repo: repo,
	}
}
func (p postService) CreatePost(ctx context.Context, Post *entity.Post) (*entity.Post, error) {
	ctx, span := otlp.Start(ctx, serviceNamePostsService, spanNamePostsService + "CreatePost")
	defer span.End()

	p.beforeRequest(nil, &Post.CreatedAt, &Post.UpdatedAt, nil)

	return p.repo.CreatePost(ctx, Post)
}
func (p postService) UpdatePost(ctx context.Context, Post *entity.PostUpdateReq) (*entity.PostUpdateReq, error) {
	ctx, span := otlp.Start(ctx, serviceNamePostsService, spanNamePostsService + "UpdatePost")
	defer span.End()

	p.beforeRequest(nil, nil, &Post.UpdatedAt, nil)

	return p.repo.UpdatePost(ctx, Post)
}
func (p postService) DeletePost(ctx context.Context, req *entity.DeleteReq) error {
	ctx, span := otlp.Start(ctx,  serviceNamePostsService, spanNamePostsService + "DeletePost")
	defer span.End()

	p.beforeRequest(nil, nil, nil, &req.DeletedAt)


	return p.repo.DeletePost(ctx, req)
}
func (p postService) GetPost(ctx context.Context, req *entity.GetReq) (*entity.Post, error) {
	ctx, span := otlp.Start(ctx, serviceNamePostsService, spanNamePostsService + "GetPost")
	defer span.End()

	return p.repo.GetPost(ctx, req.Filter)
}
func (p postService) ListPost(ctx context.Context, req *entity.ListReq) (*entity.PostListRes, error) {
	ctx, span := otlp.Start(ctx, serviceNamePostsService, spanNamePostsService + "ListPost")
	defer span.End()

	return p.repo.ListPost(ctx, req.Limit, req.Offset, req.Filter)
}
func (p postService)Search(ctx context.Context, req *entity.ListReq)(*entity.PostListRes, error){
	ctx, span := otlp.Start(ctx, serviceNamePostsService, spanNamePostsService + "Search")
	defer span.End()

	return p.repo.Search(ctx, req)
}