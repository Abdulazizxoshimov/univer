package app

import (
	"fmt"
	"net/http"
	"time"
	"univer/api"
	"univer/internal/infrastructure/clientService"
	repo "univer/internal/infrastructure/repository/postgres"
	redisrepo "univer/internal/infrastructure/repository/redisdb"
	"univer/internal/pkg/config"
	"univer/internal/pkg/logger"
	"univer/internal/pkg/otlp"
	storage "univer/internal/pkg/storage"
	"univer/internal/usecase"

	defaultrolemanager "github.com/casbin/casbin/v2/rbac/default-role-manager"
	"github.com/casbin/casbin/v2/util"
	"go.uber.org/zap"

	"github.com/casbin/casbin/v2"
)

type App struct {
	Config       config.Config
	Logger       *zap.Logger
	server       *http.Server
	DB           *storage.PostgresDB
	ShutdownOTLP func() error
	Enforcer     *casbin.Enforcer
	RedisDB      *storage.RedisDB
	User         usecase.User
	Post         usecase.Post
	Category     usecase.Category
	Comment      usecase.Comment
}

func NewApp(cfg config.Config) (*App, error) {
	// init logger
	logger, err := logger.New(cfg.LogLevel, cfg.Environment, cfg.App+".log")
	if err != nil {
		return nil, err
	}
	//init otlp collector
	shutdownOTLP, err := otlp.InitOTLPProvider(&cfg)
	if err != nil {
		return nil, err
	}
	//init redis
	redisdb, err := storage.NewRedis(&cfg)
	if err != nil {
		return nil, err
	}

	//init casbin enforcer
	enforcer, err := casbin.NewEnforcer("auth.conf", "auth.csv")
	if err != nil {
		return nil, err
	}

	// init db
	db, err := storage.New(&cfg)
	if err != nil {
		return nil, err
	}
	var (
		contextTimeout time.Duration
	)

	serviceuser := repo.NewUserRepo(db)
	userRepo := usecase.NewUserService(contextTimeout, serviceuser)

	servicecomment := repo.NewCommentRepo(db)
	commentRepo := usecase.NewCommentService(contextTimeout, servicecomment)

	servicepost := repo.NewPostRepo(db)
	postRepo := usecase.NewPostService(contextTimeout, servicepost)

	servicecategory := repo.NewCategoryRepo(db)
	categoryRepo := usecase.NewCategoryService(contextTimeout, servicecategory)

	return &App{
		Config:       cfg,
		Logger:       logger,
		DB:           db,
		ShutdownOTLP: shutdownOTLP,
		RedisDB:      redisdb,
		Enforcer:     enforcer,
		User:         userRepo,
		Post:         postRepo,
		Category:     categoryRepo,
		Comment:      &commentRepo,
	}, nil
}

func (a *App) Run() error {
	contextTimeout, err := time.ParseDuration(a.Config.Context.Timeout)
	if err != nil {
		return fmt.Errorf("error while parsing context timeout: %v", err)
	}

	service := clientService.New(a.User, a.Post, a.Comment, a.Category)

	// initialize cache
	cache := redisrepo.NewCache(a.RedisDB)

	// api init
	handler := api.NewRoute(api.RouteOption{
		Config:         a.Config,
		Logger:         a.Logger,
		ContextTimeout: contextTimeout,
		Cache:          cache,
		Enforcer:       a.Enforcer,
		Service:        service,
	})

	err = a.Enforcer.LoadPolicy()
	if err != nil {
		return err
	}
	roleManager := a.Enforcer.GetRoleManager().(*defaultrolemanager.RoleManagerImpl)

	roleManager.AddMatchingFunc("keyMatch", util.KeyMatch)
	roleManager.AddMatchingFunc("keyMatch3", util.KeyMatch3)

	// server init
	a.server, err = api.NewServer(&a.Config, handler)
	if err != nil {
		return fmt.Errorf("error while initializing server: %v", err)
	}

	return a.server.ListenAndServe()
}

func (a *App) Stop() {
	// database connection
	a.DB.Close()

	// shutdown otlp collector
	if err := a.ShutdownOTLP(); err != nil {
		a.Logger.Error("shutdown otlp collector", zap.Error(err))
	}

	// zap logger sync
	a.Logger.Sync()
}
