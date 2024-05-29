package usecase

import (
	"context"
	"log"
	"time"
	"univer/internal/entity"
	"univer/internal/infrastructure/repository"
	"univer/internal/pkg/otlp"

)

const (
	serviceNameCommentService = "commentServiceUsercase"
	spanNameCommentService    = "commentSpanUsercase"
)

type Comment interface {
	CreateComment(ctx context.Context, comment *entity.Comment) (*entity.Comment, error)
	UpdateComment(ctx context.Context, category *entity.CommentUpdateReq) (*entity.CommentUpdateReq, error)
	DeleteComment(ctx context.Context, req *entity.DeleteReq) error
	GetComment(ctx context.Context, req *entity.GetReq) (*entity.Comment, error)
	ListComment(ctx context.Context, req *entity.ListReq) (*entity.CommentListRes, error)
	CreateLike(ctx context.Context, req *entity.Like) (bool, error)
	CreateDislike(ctx context.Context, req *entity.Like) (bool, error)
}

type commentService struct {
	BaseUseCase
	repo       repository.Comment
	ctxTimeout time.Duration
}

func NewCommentService(ctxTimeout time.Duration, repo repository.Comment) commentService {
	return commentService{
		repo:       repo,
		ctxTimeout: ctxTimeout,
	}
}

func (p commentService) CreateComment(ctx context.Context, comment *entity.Comment) (*entity.Comment, error) {
	ctx, span := otlp.Start(ctx, serviceNameCommentService, spanNameCommentService+"CreateComment")
	defer span.End()

	p.beforeRequest(&comment.Id, &comment.CreatedAt, &comment.UpdatedAt, nil)

	return p.repo.CreateComment(ctx, comment)
}

func (p commentService) UpdateComment(ctx context.Context, comment *entity.CommentUpdateReq) (*entity.CommentUpdateReq, error) {
	ctx, span := otlp.Start(ctx, serviceNameCommentService, spanNameCommentService+"UpdateComment")
	defer span.End()

	p.beforeRequest(nil, nil, &comment.UpdatedAt, nil)

	return p.repo.UpdateComment(ctx, comment)
}

func (p commentService) DeleteComment(ctx context.Context, req *entity.DeleteReq) error {
	ctx, span := otlp.Start(ctx, serviceNameCommentService, spanNameCommentService+"DeleteComment")
	defer span.End()

	p.beforeRequest(nil, nil, nil, &req.DeletedAt)

	return p.repo.DeleteComment(ctx, req)
}

func (p *commentService) GetComment(ctx context.Context, req *entity.GetReq) (*entity.Comment, error) {
	ctx, span := otlp.Start(ctx, serviceNameCommentService, spanNameCommentService+"GetComment")
	defer span.End()

	return p.repo.GetComment(ctx, req.Filter)
}

func (p *commentService) ListComment(ctx context.Context, req *entity.ListReq) (*entity.CommentListRes, error) {
	ctx, span := otlp.Start(ctx, serviceNameCommentService, spanNameCommentService+"ListComment")
	defer span.End()

	return p.repo.ListComment(ctx, req.Limit, req.Offset, req.Filter)
}

func (p *commentService) CreateLike(ctx context.Context, req *entity.Like) (bool, error) {
	ctx, span := otlp.Start(ctx, serviceNameCommentService, spanNameCommentService+"CreateLike")
	defer span.End()

	status, err := p.repo.IsUnique(ctx, req.OwnerId, req.PostId, req.CommentId)
	if err != nil {
		log.Println(err)
		return false, err
	}
	if status {
		like, err := p.repo.GetLike(ctx, req)
		if err != nil {
			log.Println(err)
			return false, err
		}
		if like.Status {
			err := p.repo.DeleteLike(ctx, req)
			if err != nil {
				log.Println(err)
				return false, err
			}

			_, err = p.repo.UpdateCommentLike(ctx, req.CommentId, false)
			if err != nil {
				log.Println(err)
				return false, err
			}
			return true, err
		} else {
			_, err := p.repo.UpdateLike(ctx, req)
			if err != nil {
				log.Println(err)
				return false, err
			}
			_, err = p.repo.UpdateCommentDislike(ctx, req.CommentId, false)
			if err != nil {
				log.Println(err)
				return false, err
			}
			_, err = p.repo.UpdateCommentLike(ctx, req.CommentId, true)
			if err != nil {
				log.Println(err)
				return false, err
			}
			return true, nil
		}
	} else {
		_, err := p.repo.CreateLike(ctx, req)
		if err != nil {
			log.Println(err)
			return false, err
		}
		_, err = p.repo.UpdateCommentLike(ctx, req.CommentId, true)
		if err != nil {
			log.Println(err)
			return false, err
		}
		return true, nil
	}
}

func (p *commentService) CreateDislike(ctx context.Context, req *entity.Like) (bool, error) {
	ctx, span := otlp.Start(ctx, serviceNameCommentService, spanNameCommentService+"CreateDislike")
	defer span.End()

	status, err := p.repo.IsUnique(ctx, req.OwnerId, req.PostId, req.CommentId)
	if err != nil {
		log.Println(err)
		return false, err
	}
	if status {
		like, err := p.repo.GetLike(ctx, req)
		if err != nil {
			log.Println(err)
			return false, err
		}
		if !like.Status {
			err := p.repo.DeleteLike(ctx, req)
			if err != nil {
				log.Println(err)
				return false, err
			}
			_, err = p.repo.UpdateCommentDislike(ctx, req.CommentId, false)
			if err != nil {
				log.Println(err)
				return false, err
			}
			return true, err
		} else {
			req.Status = false
			_, err := p.repo.UpdateLike(ctx, req)
			if err != nil {
				log.Println(err)
				return false, err
			}
			_, err = p.repo.UpdateCommentLike(ctx, req.CommentId, false)
			if err != nil {
				log.Println(err)
				return false, err
			}
			_, err = p.repo.UpdateCommentDislike(ctx, req.CommentId, true)
			if err != nil {
				log.Println(err)
				return false, err
			}
			return true, nil
		}
	} else {
		_, err := p.repo.CreateLike(ctx, req)
		if err != nil {
			log.Println(err)
			return false, err
		}
		_, err = p.repo.UpdateCommentLike(ctx, req.CommentId, true)
		if err != nil {
			log.Println(err)
			return false, err
		}
		return false, nil
	}
}
