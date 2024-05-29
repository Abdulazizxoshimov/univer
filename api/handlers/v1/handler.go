package v1

import (
	"time"
	"univer/internal/infrastructure/clientService"
	repo "univer/internal/infrastructure/repository/redisdb"
	"univer/internal/pkg/config"
	tokens "univer/internal/pkg/token"

	"github.com/casbin/casbin/v2"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

type HandlerV1 struct {
	Config         config.Config
	Logger         *zap.Logger
	ContextTimeout time.Duration
	redisStorage   repo.Cache
	RefreshToken   tokens.JWTHandler
	Enforcer       *casbin.Enforcer
	Service        clientService.ServiceClient
	MinIO          *minio.Client
}

// HandlerV1Config ...
type HandlerV1Config struct {
	Config         config.Config
	Logger         *zap.Logger
	ContextTimeout time.Duration
	Redis          repo.Cache
	RefreshToken   tokens.JWTHandler
	Enforcer       *casbin.Enforcer
	Service        clientService.ServiceClient
	MinIO          *minio.Client
}

// New ...
func New(c *HandlerV1Config) *HandlerV1 {
	return &HandlerV1{
		Config:         c.Config,
		Logger:         c.Logger,
		ContextTimeout: c.ContextTimeout,
		redisStorage:   c.Redis,
		Enforcer:       c.Enforcer,
		RefreshToken:   c.RefreshToken,
		Service:        c.Service,
		MinIO:          c.MinIO,
	}
}
