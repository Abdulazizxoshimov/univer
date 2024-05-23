package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"
	"univer/api/models"
	"univer/internal/entity"
	"univer/internal/pkg/config"
	"univer/internal/pkg/image"
	regtool "univer/internal/pkg/regtool"
	tokens "univer/internal/pkg/token"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// GoogleLogin godoc
// @Summary Redirect to Google for login
// @Description Redirects the user to Google's OAuth 2.0 consent page
// @Tags auth
// @Success 303 {string} string "Redirect"
// @Router /v1/google/login [get]
func (h *HandlerV1) GoogleLogin(c *gin.Context) {
	googleConfig := config.SetupConfig()
	url := googleConfig.AuthCodeURL("RandomState")
	c.Redirect(http.StatusSeeOther, url)
	c.JSON(303, url)
}

// GoogleCallback godoc
// @Summary Handle Google callback
// @Description Handles the callback from Google OAuth 2.0, exchanges code for token and retrieves user info
// @Tags auth
// @Param state query string true "OAuth State"
// @Param code query string true "OAuth Code"
// @Success 200 {string} models.LoginResp "User info"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/google/callback [get]
func (h *HandlerV1) GoogleCallback(c *gin.Context) {
	duration, err := time.ParseDuration(h.Config.Context.Timeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	query := c.Request.URL.Query()
	state := query.Get("state")
	if state != "RandomState" {
		c.String(http.StatusUnauthorized, "state mismatch")
		return
	}

	code := query.Get("code")
	if code == "" {
		c.String(http.StatusBadRequest, "missing code")
		return
	}

	googleConfig := config.SetupConfig()

	token, err := googleConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Println("token exchange error:", err)
		c.String(http.StatusInternalServerError, "token exchange failed: %v", err)
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to get user info: %v", err)
		return
	}
	defer resp.Body.Close()

	userData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to read user info: %v", err)
		return
	}
	var body models.GoogleUser

	err = json.Unmarshal(userData, &body)
	if err != nil {
		c.JSON(303, models.Error{
			Message: err.Error(),
		})
	}
	id := uuid.New().String()
	hashpassword, err := regtool.HashPassword(body.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	h.RefreshToken = tokens.JWTHandler{
		Sub:        id,
		Role:       "user",
		SigningKey: h.Config.Token.SignInKey,
		Log:        h.Logger,
		Email:      body.Email,
	}

	access, refresh, err := h.RefreshToken.GenerateAuthJWT()

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	status, err := h.Service.User().UniqueEmail(ctx, &entity.IsUnique{
		Email: body.Email,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		return
	}
	if status.Status {
		filter := map[string]string{
			"email": body.Email,
		}
		responesUser, err := h.Service.User().GetUser(ctx, &entity.GetReq{
			Filter: filter,
		})
		if err != nil {
			c.JSON(400, models.Error{
				Message: err.Error(),
			})
		}
		c.JSON(http.StatusOK, models.UserResponse{
			Id:          responesUser.Id,
			UserName:    responesUser.Email,
			Email:       responesUser.Email,
			Bio:         responesUser.Bio,
			PhoneNumber: responesUser.PhoneNumber,
			ImageUrl:    responesUser.ImageUrl,
			Role:        responesUser.Role,
			Refresh:     refresh,
			Access:      access,
		})
		return
	}
	endpoint := h.Config.Minio.Endpoint
	accessKeyID := h.Config.Minio.AccessKeyID
	secretAccessKey := h.Config.Minio.SecretAcessKey
	bucketName := h.Config.Minio.ImageUrlUploadBucketName

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		panic(err)
	}
	err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "BucketAlreadyOwnedByYou" {

		} else {
			c.JSON(http.StatusInternalServerError, models.Error{
				Message: err.Error(),
			})
			log.Println(err.Error())
			return
		}
	}

	policy := fmt.Sprintf(`{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Principal": {
                    "AWS": ["*"]
                },
                "Action": ["s3:GetObject"],
                "Resource": ["arn:aws:s3:::%s/*"]
            }
        ]
    }`, bucketName)

	err = minioClient.SetBucketPolicy(context.Background(), bucketName, policy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	image.Image(body.Email, id)

	uploadDir := "./avatar"

	objectName := id + ".png"

	uploadPath := filepath.Join(uploadDir, objectName)

	contentType := "image/jpeg"
	_, err = minioClient.FPutObject(context.Background(), bucketName, objectName, uploadPath, minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	minioURL := fmt.Sprintf("http://%s/%s/%s", endpoint, bucketName, objectName)

	Resp, err := h.Service.User().CreateUser(ctx, &entity.User{
		Id:       id,
		UserName: body.Email,
		Email:    body.Email,
		Password: hashpassword,
		Role:     "user",
		ImageUrl: minioURL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{
		Id:       Resp.Id,
		UserName: body.Email,
		Email:    body.Email,
		Role:     "user",
		Refresh:  refresh,
		Access:   access,
		ImageUrl: minioURL,
	})
}
