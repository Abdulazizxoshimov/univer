package api

import (
	"time"

	_ "univer/api/docs"
	v1 "univer/api/handlers/v1"
	"univer/api/middleware"

	"univer/internal/infrastructure/clientService"
	redisrepo "univer/internal/infrastructure/repository/redisdb"

	"univer/internal/pkg/config"
	"univer/internal/pkg/token"

	"github.com/casbin/casbin/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

type RouteOption struct {
	Config         config.Config
	Logger         *zap.Logger
	ContextTimeout time.Duration
	Cache          redisrepo.Cache
	Enforcer       *casbin.Enforcer
	RefreshToken   token.JWTHandler
	Service        clientService.ServiceClient
}

// NewRoute
// @Title Lib-Univer
// @Description Contacs: https://t.me/Abuzada0401
// @securityDefinitions.apikey BearerAuth
// @in 			header
// @name 		Authorization
func NewRoute(option RouteOption) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	HandlerV1 := v1.New(&v1.HandlerV1Config{
		Config:         option.Config,
		Logger:         option.Logger,
		ContextTimeout: option.ContextTimeout,
		Redis:          option.Cache,
		RefreshToken:   option.RefreshToken,
		Enforcer:       option.Enforcer,
		Service:        option.Service,
	})

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"*"}
	corsConfig.AllowBrowserExtensions = true
	corsConfig.AllowMethods = []string{"*"}
	router.Use(cors.New(corsConfig))

	router.Use(middleware.Tracing)
    router.Use(middleware.CheckCasbinPermission(option.Enforcer, option.Config))

	router.Static("/media", "./media")

	apiV1 := router.Group("/v1")

	// login
	apiV1.POST("/register", HandlerV1.Register)
	apiV1.POST("/login", HandlerV1.Login)
	apiV1.POST("/forgot/:email", HandlerV1.Forgot)
	apiV1.POST("/verify", HandlerV1.VerifyOTP)
	apiV1.PUT("/reset-password", HandlerV1.ResetPassword)
	apiV1.GET("/token/:refresh", HandlerV1.Token)
	apiV1.POST("/users/verify", HandlerV1.Verify)

	//user
	apiV1.POST("/user", HandlerV1.CreateUser)
	apiV1.PUT("/user", HandlerV1.UpdateUser)
	apiV1.DELETE("/user/:id", HandlerV1.DeleteUser)
	apiV1.GET("/user/:id", HandlerV1.GetUser)
	apiV1.GET("/del/user/:id", HandlerV1.GetDelUser)
	apiV1.GET("/users", HandlerV1.ListUsers)
	apiV1.PUT("/user/profile", HandlerV1.UpdateProfile)
	apiV1.PUT("/user/password", HandlerV1.UpdatePassword)

	//post
	apiV1.POST("/post", HandlerV1.CreatePost)
	apiV1.PUT("/post", HandlerV1.UpdatePost)
	apiV1.DELETE("/post/:id", HandlerV1.DeletePost)
	apiV1.GET("/post/:id", HandlerV1.GetPost)
	apiV1.GET("/del/post/:id", HandlerV1.GetDelPost)
	apiV1.GET("/posts", HandlerV1.ListPost)
	apiV1.GET("/user/posts", HandlerV1.GetAllPostByUserId)

	// category
	apiV1.POST("/category", HandlerV1.CreateCategory)
	apiV1.PUT("/category", HandlerV1.UpdateCategory)
	apiV1.DELETE("/category/:id", HandlerV1.DeleteCategory)
	apiV1.GET("/category/:id", HandlerV1.GetCategory)
	apiV1.GET("/categories", HandlerV1.ListCategory)

	//search 
	apiV1.GET("/search", HandlerV1.Search)

	//google
	apiV1.GET("/google/callback", HandlerV1.GoogleCallback)
	apiV1.GET("google/login", HandlerV1.GoogleLogin)


	//comment
	apiV1.POST("/comment", HandlerV1.CreateComment)
	apiV1.PUT("/comment", HandlerV1.UpdateComment)
	apiV1.DELETE("/comment/:id", HandlerV1.DeleteComment)
	apiV1.GET("/comment/:id", HandlerV1.GetComment)
	apiV1.GET("/comments", HandlerV1.ListComment)
	apiV1.GET("/user/comments", HandlerV1.GetAllCommentByUserId)
	apiV1.GET("/post/comments", HandlerV1.GetAllCommentByPostId)
    apiV1.POST("/comment/dislike", HandlerV1.CreateDisLike)
	apiV1.POST("/comment/like", HandlerV1.CreateLike)

	


	url := ginSwagger.URL("swagger/doc.json")
	apiV1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	return router
}
