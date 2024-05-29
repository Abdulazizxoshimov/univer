package repository

import (
	"context"
	"univer/internal/entity"
)

type Post interface {
	CreatePost(ctx context.Context, post *entity.Post) (*entity.Post, error)
	UpdatePost(ctx context.Context, post *entity.PostUpdateReq) (*entity.PostUpdateReq, error)
	DeletePost(ctx context.Context, req *entity.DeleteReq) error
	GetPost(ctx context.Context, params map[string]string) (*entity.Post, error)
	ListPost(ctx context.Context, limit int, offset int, filter map[string]string) (*entity.PostListRes, error)
	Search(ctx context.Context, req *entity.ListReq)(*entity.PostListRes, error)
	CheckUnique(ctx context.Context, UserId, PostId string) (bool, error)
	CreateViews(ctx context.Context, userId, postId string) (bool, error)
	UpdateViews(ctx context.Context, postId string)(bool, error)
}
