package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"univer/api/models"
	"univer/internal/entity"
	"univer/internal/pkg/config"
	"univer/internal/pkg/image"
	regtool "univer/internal/pkg/regtool"
	tokens "univer/internal/pkg/token"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
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
	ctx, cancel := context.WithTimeout(context.Background(), h.Config.Context.Timeout)
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
	filter := map[string]string{
		"email": body.Email,
	}

	status, err := h.Service.User().CheckUnique(ctx, &entity.GetReq{
		Filter: filter,
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
		h.RefreshToken = tokens.JWTHandler{
			Sub:        responesUser.Id,
			Role:       responesUser.Role,
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

	image, err := image.Image(body.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	objectName := id + ".png"
	contentType := c.ContentType()

	var buf bytes.Buffer
	err = png.Encode(&buf, image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	reader := bytes.NewReader(buf.Bytes())
	if h.MinIO == nil {
		return
	}
	_, err = h.MinIO.PutObject(ctx, h.Config.Minio.ImageUrlUploadBucketName, objectName, reader, int64(reader.Len()), minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	minioURL := fmt.Sprintf("http://%s/%s/%s", h.Config.Minio.Endpoint, h.Config.Minio.ImageUrlUploadBucketName, objectName)

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
