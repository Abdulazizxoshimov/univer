package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"univer/api/models"
	"univer/internal/entity"
	"univer/internal/pkg/image"
	regtool "univer/internal/pkg/regtool"
	tokens "univer/internal/pkg/token"
	validation "univer/internal/pkg/validation"

	govalidator "github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/encoding/protojson"
)

// @Summary 		Register
// @Description 	Api for register user
// @Tags 			registration
// @Accept 			json
// @Produce 		json
// @Param 			User body models.UserRegister true "Register User"
// @Success 		200 {object} models.User
// @Failure 		400 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/register [POST]
func (h *HandlerV1) Register(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(7))
	defer cancel()

	var (
		body        models.UserRegister
		jspbMarshal protojson.MarshalOptions
	)
	jspbMarshal.UseProtoNames = true

	err := c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}
	valid := govalidator.IsEmail(body.Email)
	if !valid {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "Bad email",
		})
		log.Println(err)
		return
	}

	body.Email, err = validation.EmailValidation(body.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	status := validation.PasswordValidation(body.Password)
	if !status {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "Password should be 8-20 characters long and contain at least one lowercase letter, one uppercase letter, and one digit",
		})
		log.Println(err)
		return
	}

	IsPhoneNumber := validation.PhoneUz(body.PhoneNumber)
	if !IsPhoneNumber {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "Incorrect telephone number!!!",
		})
		log.Println(err)
		return
	}

	exists, err := h.Service.User().UniqueEmail(ctx, &entity.IsUnique{
		Email: body.Email,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	if exists.Status {
		c.JSON(http.StatusConflict, models.Error{
			Message: "This email already in use:",
		})
		return
	}

	radomNumber, err := regtool.SendCodeGmail(body.Email, "ClothesStore\n", "./internal/pkg/regtool/emailotp.html", h.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	err = h.redisStorage.Set(ctx, radomNumber, body, time.Second*300)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Response: "Verification code sent  your email",
	})
}

// @Summary            Verify
// @Description        Api for verify register
// @Tags               registration
// @Accept             json
// @Produce            json
// @Param              email query string true "email"
// @Param              code query string true "code"
// @Success            201 {object} models.UserResponse
// @Failure            400 {object} models.Error
// @Failure            500 {object} models.Error
// @Router             /v1/users/verify [post]
func (h *HandlerV1) Verify(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(7))
	defer cancel()

	code := c.Query("code")
	email := c.Query("email")

	userData, err := h.redisStorage.Get(ctx, code)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}
	var user models.User

	err = json.Unmarshal(userData, &user)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}
	if user.Email != email {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "The email did not match ",
		})
		log.Println(err)
		return
	}

	id := uuid.NewString()

	h.RefreshToken = tokens.JWTHandler{
		Sub:        id,
		Role:       "user",
		SigningKey: h.Config.Token.SignInKey,
		Log:        h.Logger,
		Email:      user.Email,
	}

	access, refresh, err := h.RefreshToken.GenerateAuthJWT()
	if err != nil {
		c.JSON(http.StatusConflict, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	hashPassword, err := regtool.HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusBadGateway, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	claims, err := tokens.ExtractClaim(access, []byte(h.Config.Token.SignInKey))
	if err != nil {
		c.JSON(http.StatusBadGateway, models.Error{
			Message: err.Error(),
		})
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
	image.Image(user.Email, id)

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

	_, err = h.Service.User().CreateUser(ctx, &entity.User{
		Id:           id,
		UserName:     user.UserName,
		Email:        user.Email,
		Password:     hashPassword,
		ImageUrl:     minioURL,
		PhoneNumber:  user.PhoneNumber,
		RefreshToken: refresh,
		Role:         cast.ToString(claims["role"]),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}
	if err := os.Remove(filepath.Join(uploadPath)); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	respUser := &models.UserResponse{
		Id:          id,
		UserName:    user.UserName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		ImageUrl:    minioURL,
		Role:        cast.ToString(claims["role"]),
		Access:      access,
		Refresh:     refresh,
	}

	c.JSON(http.StatusCreated, respUser)
}

// @Summary 		Login
// @Description 	Api for login user
// @Tags 			registration
// @Accept 			json
// @Produce 		json
// @Param 			login body models.Login true "Login Model"
// @Success 		200 {object} models.UserResponse
// @Failure 		400 {object} models.Error
// @Failure 		404 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/login [POST]
func (h *HandlerV1) Login(c *gin.Context) {
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

	var body models.Login

	err = c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	var filter map[string]string
	if govalidator.IsEmail(body.UserNameOrEmail) {
		filter = map[string]string{
			"email": body.UserNameOrEmail,
		}
	} else {
		filter = map[string]string{
			"username": body.UserNameOrEmail,
		}
	}

	response, err := h.Service.User().GetUser(
		ctx, &entity.GetReq{
			Filter: filter,
		})
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	if !(regtool.CheckHashPassword(body.Password, response.Password)) {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "Incorrect Password",
		})
		return
	}

	h.RefreshToken = tokens.JWTHandler{
		Sub:        response.Id,
		Role:       response.Role,
		SigningKey: h.Config.Token.SignInKey,
		Log:        h.Logger,
		Email:      response.Email,
	}

	access, refresh, err := h.RefreshToken.GenerateAuthJWT()

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	respUser := &models.UserResponse{
		Id:          response.Id,
		UserName:    response.UserName,
		Email:       response.Email,
		PhoneNumber: response.PhoneNumber,
		ImageUrl:    response.ImageUrl,
		Bio:         response.Bio,
		Role:        response.Role,
		Refresh:     refresh,
		Access:      access,
	}
	_, err = h.Service.User().UpdateRefresh(ctx, &entity.UpdateRefresh{
		UserID:       response.Id,
		RefreshToken: refresh,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	c.JSON(http.StatusOK, respUser)
}

// @Summary 		Forget Password
// @Description 	Api for sending otp
// @Tags 			registration
// @Accept 			json
// @Produce 		json
// @Param 			email path string true "Email"
// @Success 		200 {object} string
// @Failure 		400 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/forgot/{email} [POST]
func (h *HandlerV1) Forgot(c *gin.Context) {
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

	email := c.Param("email")

	email, err = validation.EmailValidation(email)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	status, err := h.Service.User().UniqueEmail(ctx, &entity.IsUnique{
		Email: email,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	if !status.Status {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "This user is not registered",
		})
		return
	}

	radomNumber, err := regtool.SendCodeGmail(email, "ClothesStore\n", "./internal/pkg/regtool/forgotpassword.html", h.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	if err := h.redisStorage.Set(ctx, radomNumber, cast.ToString(email), time.Second*300); err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	c.JSON(http.StatusOK, "We have sent otp your email")
}

// @Summary 		Verify OTP
// @Description 	Api for verify user
// @Tags 			registration
// @Accept 			json
// @Produce 		json
// @Param 			email query string true "Email"
// @Param 			otp query string true "OTP"
// @Success 		200 {object} bool
// @Failure 		400 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/verify [POST]
func (h *HandlerV1) VerifyOTP(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(7))
	defer cancel()

	otp := c.Query("otp")
	email := c.Query("email")

	userData, err := h.redisStorage.Get(ctx, otp)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}
	var redisEmail string

	err = json.Unmarshal(userData, &redisEmail)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	if redisEmail != email {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "The email did not match",
		})
		log.Println("The email did not match")
		return
	}

	c.JSON(http.StatusCreated, true)
}

// @Summary 		Reset Password
// @Description 	Api for reset password
// @Tags 			registration
// @Accept 			json
// @Produce 		json
// @Param 			User body models.ResetPassword true "Reset Password"
// @Success 		200 {object} bool
// @Failure 		400 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/reset-password [PUT]
func (h *HandlerV1) ResetPassword(c *gin.Context) {
	var (
		body models.ResetPassword
	)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*time.Duration(7))
	defer cancel()

	err := c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	status := validation.PasswordValidation(body.NewPassword)
	if !status {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "Password should be 8-20 characters long and contain at least one lowercase letter, one uppercase letter, one symbol, and one digit",
		})
		log.Println(err)
		return
	}

	user, err := h.Service.User().GetUser(ctx, &entity.GetReq{
		Filter: map[string]string{
			"email": body.Email,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	hashPassword, err := regtool.HashPassword(body.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadGateway, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	responseStatus, err := h.Service.User().UpdatePassword(ctx, &entity.UpdatePassword{
		UserID:      user.Id,
		NewPassword: hashPassword,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	if !responseStatus.Status {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: "Password doesn't updated",
		})
		log.Println("Password doesn't updated")
		return
	}

	c.JSON(http.StatusOK, true)
}

// @Summary 		New Token
// @Description 	Api for updated acces token
// @Tags 			registration
// @Accept 			json
// @Produce 		json
// @Param 			refresh path string true "Refresh Token"
// @Success 		200 {object} models.TokenResp
// @Failure 		400 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		409 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/token/{refresh} [GET]
func (h *HandlerV1) Token(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(7))
	defer cancel()

	RToken := c.Param("refresh")

	user, err := h.Service.User().GetUser(ctx, &entity.GetReq{
		Filter: map[string]string{
			"refresh_token": RToken,
		},
	})

	if err != nil {
		c.JSON(500, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	resclaim, err := tokens.ExtractClaim(RToken, []byte(h.Config.Token.SignInKey))
	if err != nil {
		c.JSON(500, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	Now_time := time.Now().Unix()
	exp := (resclaim["exp"])
	if exp.(float64)-float64(Now_time) > 0 {
		h.RefreshToken = tokens.JWTHandler{
			Sub:        user.Id,
			Role:       user.Role,
			SigningKey: h.Config.Token.SignInKey,
			Log:        h.Logger,
			Email:      user.Email,
		}

		access, refresh, err := h.RefreshToken.GenerateAuthJWT()
		if err != nil {
			c.JSON(http.StatusConflict, models.Error{
				Message: err.Error(),
			})
			log.Println(err)
			return
		}

		_, err = h.Service.User().UpdateRefresh(ctx, &entity.UpdateRefresh{
			UserID:       user.Id,
			RefreshToken: refresh,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, models.Error{
				Message: err.Error(),
			})
			log.Println(err)
			return
		}

		respUser := &models.TokenResp{
			ID:      user.Id,
			Role:    user.Role,
			Refresh: refresh,
			Access:  access,
		}

		c.JSON(http.StatusCreated, respUser)
	} else {
		c.JSON(http.StatusUnauthorized, models.Error{
			Message: "refresh token expired",
		})
		log.Println("refresh token expired")
		return
	}
}
