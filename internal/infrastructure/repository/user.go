package repository

import (
	"context"
	"univer/internal/entity"
)


type User interface{
	Create(ctx context.Context, req *entity.User)(*entity.User, error)
	Get(ctx context.Context, params map[string]string)(*entity.User, error)
	Update(ctx context.Context, req *entity.User)(*entity.User, error)
	List(ctx context.Context, page, offset int, filter map[string]string)(*entity.ListUserRes, error)
	Delete(ctx context.Context, Filter *entity.DeleteReq)error
	CheckUnique(ctx context.Context, filter *entity.GetReq) (bool, error)
	UpdateRefresh(ctx context.Context, request *entity.UpdateRefresh) (*entity.Response, error)
	UpdatePassword(ctx context.Context, request *entity.UpdatePassword) (*entity.Response, error)
	UpdateProfile(ctx context.Context, request *entity.UpdateProfile) (*entity.Response, error)
	DeleteProfile(ctx context.Context, id string)error
	UpdateToPremium(ctx context.Context, id string) (*entity.Response, error)
}