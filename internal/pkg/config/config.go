package config

import (
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type webAddress struct {
	Host string
	Port string
}

type Config struct {
	App         string
	Environment string
	LogLevel    string
	Server      struct {
		Host         string
		Port         string
		ReadTimeout  string
		WriteTimeout string
		IdleTimeout  string
	}

	Context struct {
		Timeout string
	}

	DB struct {
		Host     string
		Port     string
		Name     string
		User     string
		Password string
		SslMode  string
	}
	Redis struct {
		Host     string
		Port     string
		Password string
		Name     string
		Time     time.Time
	}
	Token struct {
		Secret     string
		AccessTTL  time.Duration
		RefreshTTL time.Duration
		SignInKey  string
	}
	Minio struct {
		Endpoint                 string
		AccessKeyID              string
		SecretAcessKey           string
		Location                 string
		ImageUrlUploadBucketName string
		FileUploadBucketName     string
	}
	SMTP struct {
		Email         string
		EmailPassword string
		SMTPPort      string
		SMTPHost      string
	}

	OTLPCollector webAddress
}

func NewConfig() (*Config, error) {
	var config Config

	// general configuration
	config.App = getEnv("APP", "app")
	config.Environment = getEnv("ENVIRONMENT", "develop")
	config.LogLevel = getEnv("LOG_LEVEL", "debug")

	// server configuration
	config.Server.Host = getEnv("SERVER_HOST", "app")
	config.Server.Port = getEnv("SERVER_PORT", ":7777")
	config.Server.ReadTimeout = getEnv("SERVER_READ_TIMEOUT", "10s")
	config.Server.WriteTimeout = getEnv("SERVER_WRITE_TIMEOUT", "10s")
	config.Server.IdleTimeout = getEnv("SERVER_IDLE_TIMEOUT", "120s")

	//context configuration
	config.Context.Timeout = getEnv("CONTEXT_TIMEOUT", "30s")

	// db configuration
	config.DB.Host = getEnv("POSTGRES_HOST", "postgres")
	config.DB.Port = getEnv("POSTGRES_PORT", "5432")
	config.DB.User = getEnv("POSTGRES_USER", "postgres")
	config.DB.Password = getEnv("POSTGRES_PASSWORD", "4444")
	config.DB.SslMode = getEnv("POSTGRES_SSLMODE", "disable")
	config.DB.Name = getEnv("POSTGRES_DATABASE", "univerdb")

	// access ttl parse
	accessTTl, err := time.ParseDuration(getEnv("TOKEN_ACCESS_TTL", "3h"))
	if err != nil {
		return nil, err
	}
	// refresh ttl parse
	refreshTTL, err := time.ParseDuration(getEnv("TOKEN_REFRESH_TTL", "24h"))
	if err != nil {
		return nil, err
	}
	config.Token.AccessTTL = accessTTl
	config.Token.RefreshTTL = refreshTTL
	config.Token.SignInKey = getEnv("TOKEN_SIGNIN_KEY", "debug")

	// otlp collector configuration
	config.OTLPCollector.Host = getEnv("OTLP_COLLECTOR_HOST", "otel-collector")
	config.OTLPCollector.Port = getEnv("OTLP_COLLECTOR_PORT", ":4318")

	// redis configuration
	config.Redis.Host = getEnv("REDIS_HOST", "redisdb")
	config.Redis.Port = getEnv("REDIS_PORT", "6379")
	config.Redis.Password = getEnv("REDIS_PASSWORD", "")
	config.Redis.Name = getEnv("REDIS_DATABASE", "0")

	//smtp confifuration
	config.SMTP.Email = getEnv("SMTP_EMAIL", "theuniver77@gmail.com")
	config.SMTP.EmailPassword = getEnv("SMTP_EMAIL_PASSWORD", "fywqgrsyhvybjyxa")
	config.SMTP.SMTPPort = getEnv("SMTP_PORT", "587")
	config.SMTP.SMTPHost = getEnv("SMTP_HOST", "smtp.gmail.com")

	//minIO configuration
	config.Minio.AccessKeyID = getEnv("ACCES_KEY", "xoshimov")
	config.Minio.SecretAcessKey = getEnv("SECRET_ACCES_KEY", "xoshimov")
	config.Minio.Endpoint = getEnv("ENDPOINT", "127.0.0.1:9000")
	config.Minio.FileUploadBucketName = getEnv("FILE_UPLOAD_BUCKET_NAME", "univer")
	config.Minio.ImageUrlUploadBucketName = getEnv("IMAGE_URL_UPLOAD_BUCKET_NAME", "univer-image")

	
	return &config, nil
}

func SetupConfig() *oauth2.Config {

	conf := &oauth2.Config{
		ClientID:     getEnv("CLIENT_ID", "380746396734-29esjuel9n313k7g1di4ccj9mb5q0pu2.apps.googleusercontent.com"),
		ClientSecret: getEnv("CLIENT_SECRET", "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"),
		RedirectURL:  getEnv("REDIRECT_URL", "http://localhost:7777/v1/google/callback"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	return conf
}

func getEnv(key string, defaultVaule string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	}
	return defaultVaule
}
