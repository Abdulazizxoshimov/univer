package repository

import (
	"context"
	"univer/internal/entity"
)



type Comment interface{
	CreateComment(ctx context.Context, comment *entity.Comment) (*entity.Comment, error)
	UpdateComment(ctx context.Context, category *entity.CommentUpdateReq) (*entity.CommentUpdateReq, error)
	DeleteComment(ctx context.Context, req *entity.DeleteReq) error
	GetComment(ctx context.Context, params map[string]string) (*entity.Comment, error)
	ListComment(ctx context.Context, limit int, offset int, params map[string]string) (*entity.CommentListRes, error)
	UpdateLike(ctx context.Context, req *entity.Like) (bool, error)
	UpdateCommentLike(ctx context.Context, id string, status bool)(bool, error)
	UpdateCommentDislike(ctx context.Context, id string, status bool)(bool, error)
	GetLike(ctx context.Context, req *entity.Like) (*entity.Like, error)
	IsUnique(ctx context.Context, OwnerId, PostId, CommentId string) (bool, error)
	CreateLike(ctx context.Context, req *entity.Like) (bool, error)
	DeleteLike(ctx context.Context, req *entity.Like) error
}