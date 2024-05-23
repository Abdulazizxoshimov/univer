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
	serviceNameUserService = "userServiceUsecase"
	spanNameUserService    = "userSpanUsecase"
)

type User interface {
	CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	DeleteUser(ctx context.Context, req *entity.DeleteReq) error
	GetUser(ctx context.Context, filter *entity.GetReq) (*entity.User, error)
	ListUser(ctx context.Context, req *entity.ListReq) (*entity.ListUserRes, error)
	UniqueEmail(ctx context.Context, req *entity.IsUnique) (*entity.Response, error)
	UpdateRefresh(ctx context.Context, request *entity.UpdateRefresh) (*entity.Response, error)
	UpdatePassword(ctx context.Context, request *entity.UpdatePassword) (*entity.Response, error)
	UpdateProfile(ctx context.Context, request *entity.UpdateProfile) (*entity.Response, error)
	DeleteProfile(ctx context.Context,  id string)error
}

type userService struct {
	BaseUseCase
	repo       repository.User
	ctxTimeout time.Duration
}

func NewUserService(ctxTimeout time.Duration, repo repository.User) userService {
	return userService{
		ctxTimeout: ctxTimeout,
		repo:       repo,
	}
}

func (u userService) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService + "CreateUser")
	defer span.End()

	u.beforeRequest(nil, &user.CreatedAt, &user.UpdatedAt, nil)

	return u.repo.Create(ctx, user)
}

func (u userService) UpdateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService + "UpdateUser")
	defer span.End()

	u.beforeRequest(nil, nil, &user.UpdatedAt, nil)

	return u.repo.Update(ctx, user)
}

func (u userService) DeleteUser(ctx context.Context, req *entity.DeleteReq) error {
	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService +  "DeleteUser")
	defer span.End()
	u.beforeRequest(nil, nil, nil, &req.DeletedAt)

	return u.repo.Delete(ctx, req)
}

func (u userService) GetUser(ctx context.Context, filter *entity.GetReq) (*entity.User, error) {
	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService + "GetUser")
	defer span.End()

	return u.repo.Get(ctx, filter.Filter)
}

func (u userService) ListUser(ctx context.Context, req *entity.ListReq) (*entity.ListUserRes, error) {
	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService + "ListUsers")
	defer span.End()

	return u.repo.List(ctx, req.Limit, req.Offset, req.Filter)
}

func (u userService) UniqueEmail(ctx context.Context, req *entity.IsUnique) (*entity.Response, error) {
	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService +"UniqueEmail")
	defer span.End()

	status, err := u.repo.UniqueEmail(ctx, req.Email)
	if err != nil{
		log.Println(err.Error())
		return &entity.Response{Status: false}, err
	}
	return &entity.Response{Status: status}, nil
}

func (u userService) UpdateRefresh(ctx context.Context, request *entity.UpdateRefresh) (*entity.Response, error) {
	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService + "UpdateRefresh")
	defer span.End()

	return u.repo.UpdateRefresh(ctx, request)
}

func (u userService) UpdatePassword(ctx context.Context, request *entity.UpdatePassword) (*entity.Response, error) {
	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService + "UpdatePassword")
	defer span.End()

	return u.repo.UpdatePassword(ctx, request)
}

func (u userService) UpdateProfile(ctx context.Context, request *entity.UpdateProfile) (*entity.Response, error) {
	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService + "UpdatePassword")
	defer span.End()

	return u.repo.UpdateProfile(ctx, request)
}

func (u userService)DeleteProfile(ctx context.Context, id string)error {
	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService + "DeleteProfile")
	defer span.End()

	return u.repo.DeleteProfile(ctx, id)
}