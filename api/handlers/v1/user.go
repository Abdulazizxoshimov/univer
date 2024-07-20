package v1

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"univer/api/models"
	"univer/internal/entity"
	regtool "univer/internal/pkg/regtool"
	"univer/internal/pkg/validation"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

// @Security  		BearerAuth
// @Summary   		Create User
// @Description 	Api for create a new user
// @Tags 			users
// @Accept 			json
// @Produce 		json
// @Param 			user body models.UserRegister true "Create User Model"
// @Success 		201 {object} models.CreateResponse
// @Failure 		400 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/user [POST]
func (h *HandlerV1) CreateUser(c *gin.Context) {
	var (
		body        models.UserRegister
	)

	ctx, cancel := context.WithTimeout(context.Background(), h.Config.Context.Timeout)
	defer cancel()

	err := c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	body.Email, err = validation.EmailValidation(body.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
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
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "Email already used",
		})
		return
	}

	statusPassword := validation.PasswordValidation(body.Password)
	if !statusPassword {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: models.WeakPasswordMessage,
		})
		log.Println(models.WeakPasswordMessage)
		return
	}

	hashpassword, err := regtool.HashPassword(body.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	if !validation.ValidateUsername(body.UserName) {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "Invalid Username",
		})
	}

	userServiceCreateResponse, err := h.Service.User().CreateUser(ctx, &entity.User{
		Id:       uuid.New().String(),
		UserName: body.UserName,
		Email:    body.Email,
		Password: hashpassword,
		Role:     "user",
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusCreated, models.CreateResponse{
		Id: userServiceCreateResponse.Id,
	})
}

// @Security  		BearerAuth
// @Summary   		Update User
// @Description 	Api for update a user
// @Tags 			users
// @Accept 			json
// @Produce 		json
// @Param 			user body models.UpdateReq true "Update User Model"
// @Success 		200 {object} models.User
// @Failure 		400 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/user [PUT]
func (h *HandlerV1) UpdateUser(c *gin.Context) {
	var (
		body        models.UpdateReq
	)

	ctx, cancel := context.WithTimeout(context.Background(), h.Config.Context.Timeout)
	defer cancel()

	err := c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	body.Email, err = validation.EmailValidation(body.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	filter := map[string]string{
		"id": body.Id,
	}
	user, err := h.Service.User().GetUser(ctx, &entity.GetReq{Filter: filter})
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
	}
	mp := map[string]string{
		"email": body.Email,
	}
	if user.Email != body.Email {
		status, err := h.Service.User().CheckUnique(ctx, &entity.GetReq{
			Filter: mp,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, models.Error{
				Message: err.Error(),
			})
			log.Println(err.Error())
			return
		}
		if status.Status {
			c.JSON(http.StatusBadRequest, models.Error{
				Message: "email already used",
			})
			return
		}
	}

	if body.PhoneNumber != "" {
		status := validation.PhoneUz(body.PhoneNumber)
		if !status {
			c.JSON(http.StatusBadRequest, models.Error{
				Message: "phone number is invalid",
			})
			log.Println("phone number is invalid")
			return
		}
	}

	updatedUser, err := h.Service.User().UpdateUser(ctx, &entity.User{
		Id:          body.Id,
		UserName:    body.UserName,
		Email:       body.Email,
		PhoneNumber: body.PhoneNumber,
		Bio:         body.Bio,
		Role:        "user",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, models.User{
		Id:          updatedUser.Id,
		UserName:    updatedUser.UserName,
		Email:       updatedUser.Email,
		PhoneNumber: updatedUser.PhoneNumber,
	})
}

// @Security  		BearerAuth
// @Summary   		Delete User
// @Description 	Api for delete a user
// @Tags 			users
// @Accept 			json
// @Produce 		json
// @Param 			id path string true "User ID"
// @Success 		200 {object} bool
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/user/{id} [DELETE]
func (h *HandlerV1) DeleteUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), h.Config.Context.Timeout)
	defer cancel()

	userID := c.Param("id")

	user, err := h.Service.User().GetUser(ctx, &entity.GetReq{
		Filter: map[string]string{
			"id": userID,
		},
	})
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	if user.Role == "admin" {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "Wrong request",
		})
		return
	}

	err = h.Service.User().DeleteUser(ctx, &entity.DeleteReq{
		Id: userID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, true)
}

// @Security  		BearerAuth
// @Summary   		Get User
// @Description 	Api for getting a user
// @Tags 			users
// @Accept 			json
// @Produce 		json
// @Param 			id path string true "User ID"
// @Success 		200 {object} models.UserResponse
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/user/{id} [GET]
func (h *HandlerV1) GetUser(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), h.Config.Context.Timeout)
	defer cancel()

	userID := c.Param("id")

	filter := make(map[string]string)

	if govalidator.IsEmail(userID) {
		filter["email"] = userID
	} else if validation.ValidateUUID(userID) {
		filter["id"] = userID
	} else {
		filter["username"] = userID
	}

	response, err := h.Service.User().GetUser(ctx, &entity.GetReq{
		Filter: filter,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{
		Id:          userID,
		UserName:    response.UserName,
		Email:       response.Email,
		PhoneNumber: response.PhoneNumber,
		Bio:         response.Bio,
		ImageUrl:    response.ImageUrl,
		Refresh:     response.RefreshToken,
		Role:        response.Role,
	})
}

// @Security  		BearerAuth
// @Summary   		Get  Delete User
// @Description 	Api for getting a deleted user
// @Tags 			users
// @Accept 			json
// @Produce 		json
// @Param 			id path string true "User ID"
// @Success 		200 {object} models.UserResponse
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/del/user/{id} [GET]
func (h *HandlerV1) GetDelUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), h.Config.Context.Timeout)
	defer cancel()

	userID := c.Param("id")
	filter := map[string]string{
		"del": userID,
	}
	response, err := h.Service.User().GetUser(ctx, &entity.GetReq{
		Filter: filter,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{
		Id:          userID,
		UserName:    response.UserName,
		Email:       response.Email,
		PhoneNumber: response.PhoneNumber,
		ImageUrl:    response.Email,
		Bio:         response.Bio,
		Role:        response.Role,
		Refresh:     response.RefreshToken,
	})
}

// @Security  		BearerAuth
// @Summary   		List User
// @Description 	Api for getting list user
// @Tags 			users
// @Accept 			json
// @Produce 		json
// @Param 			page query string true "Page"
// @Param 			limit query string true "Limit"
// @Success 		200 {object} models.ListUser
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/users [GET]
func (h *HandlerV1) ListUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), h.Config.Context.Timeout)
	defer cancel()
	page := c.Query("page")
	limit := c.Query("limit")
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	offset := (pageInt - 1) * limitInt
	filter := map[string]string{
		"role": "user",
	}
	listUsers, err := h.Service.User().ListUser(ctx, &entity.ListReq{
		Offset: offset,
		Limit:  limitInt,
		Filter: filter,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	var users []models.UserResponse
	for _, user := range listUsers.User {
		users = append(users, models.UserResponse{
			Id:          user.Id,
			UserName:    user.UserName,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			Bio:         user.Bio,
			ImageUrl:    user.ImageUrl,
			Refresh:     user.RefreshToken,
			Role:        user.Role,
		})
	}

	c.JSON(http.StatusOK, models.ListUser{
		User:  users,
		Total: uint64(listUsers.TotalCount),
	})
}

// @Security        BearerAuth
// @Summary         Update Profile
// @Description     Api for updating user's profile
// @Tags            users
// @Accept          json
// @Produce         json
// @Param 			file formData file true "File"
// @Success 		200 {object} models.Response
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/user/profile [PUT]
func (h *HandlerV1) UpdateProfile(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), h.Config.Context.Timeout)
	defer cancel()

	userId, statusCode := GetIdFromToken(c.Request, &h.Config)
	if statusCode != 0 {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "oops something went wrong",
		})
	}

	file := &models.File{}
	err := c.ShouldBind(&file)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	if file.File.Size > 10<<20 {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "File size cannot be larger than 10 MB",
		})
		return
	}

	ext := filepath.Ext(file.File.Filename)
	ext = strings.ToLower(ext)
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".tiff": true,
		".svg":  true,
	}
	if !allowedExtensions[ext] {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "Only .jpg, .jpeg, .png, .gif, .bmp, .tiff, and .svg format files are accepted"})
		return
	}

	fileHeader, err := file.File.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}
	defer fileHeader.Close()

	objectName := userId + ext
	contentType := c.ContentType()
	_, err = h.MinIO.PutObject(ctx, h.Config.Minio.ImageUrlUploadBucketName, objectName, fileHeader, file.File.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	minioURL := fmt.Sprintf("http://localhost:9000/%s/%s", h.Config.Minio.ImageUrlUploadBucketName, objectName)

	_, err = h.Service.User().UpdateProfile(ctx, &entity.UpdateProfile{
		Id:       userId,
		ImageUrl: minioURL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Response: minioURL,
	})
}

// @Security        BearerAuth
// @Summary         Update Password
// @Description     Api for updating user's password
// @Tags            users
// @Accept          json
// @Produce         json
// @Param 			user body models.UpdatePasswordReq true "Update User Password"
// @Success 		200 {object} models.Response
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/user/password [PUT]
func (h *HandlerV1) UpdatePassword(c *gin.Context) {
	var (
		body        models.UpdatePasswordReq
	)

	ctx, cancel := context.WithTimeout(context.Background(), h.Config.Context.Timeout)
	defer cancel()

	err := c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	_, err = h.Service.User().UpdatePassword(ctx, &entity.UpdatePassword{
		UserID:      body.Id,
		NewPassword: body.Password,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Response: "Your profile has changed succesfully",
	})
}

// @Security        BearerAuth
// @Summary         Update To Premium
// @Description     Api for updating user's role
// @Tags            users
// @Produce         json
// @Param           id path string true "Id"
// @Success 		200 {object} models.Response
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/user/premium/{id} [PUT]
func (h *HandlerV1) UpdateToPremium(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), h.Config.Context.Timeout)
	defer cancel()

	id := c.Param("id")

	_, err := h.Service.User().UpdateToPremium(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Response: "Your profile has changed succesfully",
	})
}
