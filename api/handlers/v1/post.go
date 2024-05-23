package v1

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"univer/api/models"
	"univer/internal/entity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"google.golang.org/protobuf/encoding/protojson"
)

// @Security  		BearerAuth
// @Summary   		Create Post
// @Description 	Api for create a new Post
// @Tags 			post
// @Accept 			multipart/form-data
// @Produce 		json
// @Param 			theme query string true "Theme"
// @Param 			science query string true "Science"
// @Param 			id query string true "Category Id"
// @Param 			file formData file true "File"
// @Success 		201 {object} models.CreateResponse
// @Failure 		400 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/post [POST]
func (h *HandlerV1) CreatePost(c *gin.Context) {
	endpoint := h.Config.Minio.Endpoint
	accessKeyID := h.Config.Minio.AccessKeyID
	secretAccessKey := h.Config.Minio.SecretAcessKey
	bucketName := h.Config.Minio.FileUploadBucketName
	
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

	var (
		file        models.File
		body        models.PostCreate
		jspbMarshal protojson.MarshalOptions
	)
	jspbMarshal.UseProtoNames = true

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

	body.CategoryId = c.Query("id")
	body.Science = c.Query("science")
	body.Theme = c.Query("theme")

	err = c.ShouldBind(&file)
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
	allowedExtensions := map[string]bool{
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".ppt":  true,
		".pptx": true,
		".xls":  true,
		".xlsx": true,
		".xlsm": true,
	}
	if !allowedExtensions[ext] {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "Only .pdf, .doc, .docx, .ppt, and .pptx format files are accepted"})
		return
	}

	uploadDir := "./media"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, os.ModePerm)
	}

	id := uuid.New().String()

	newFilename := id + ext
	uploadPath := filepath.Join(uploadDir, newFilename)

	if err := c.SaveUploadedFile(file.File, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	objectName := body.Theme + ext
	contentType := c.ContentType()
	_, err = minioClient.FPutObject(ctx, bucketName, objectName, uploadPath, minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}
	if err := os.Remove(uploadPath); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	minioURL := fmt.Sprintf("http://%s/%s/%s", endpoint, bucketName, objectName)

	userId, statusCode := GetIdFromToken(c.Request, &h.Config)
	if statusCode != 0 {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "oops something went wrong",
		})
	}
	post, err := h.Service.Post().CreatePost(ctx, &entity.Post{
		Id:         id,
		UserId:     userId,
		Theme:      body.Theme,
		Path:       minioURL,
		Science:    body.Science,
		CategoryId: body.CategoryId,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusCreated, models.CreateResponse{
		Id: post.Id,
	})
}

// @Security  		BearerAuth
// @Summary   		Update Post
// @Description 	Api for update a post
// @Tags 			post
// @Accept 			json
// @Produce 		json
// @Param 			post body models.PostUpdateReq true "Update Post Model"
// @Success 		200 {object} models.PostUpdateReq
// @Failure 		400 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/post [PUT]
func (h *HandlerV1) UpdatePost(c *gin.Context) {
	var (
		body        models.PostUpdateReq
		jspbMarshal protojson.MarshalOptions
	)
	jspbMarshal.UseProtoNames = true

	duration, err := time.ParseDuration(h.Config.Context.Timeout)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	err = c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	post, err := h.Service.Post().UpdatePost(ctx, &entity.PostUpdateReq{
		Id:         body.Id,
		Theme:      body.Theme,
		Science:    body.Science,
		CategoryId: body.CategoryId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, models.PostUpdateReq{
		Id:         post.Id,
		Theme:      post.Theme,
		Science:    post.Science,
		CategoryId: post.CategoryId,
	})
}

// @Security  		BearerAuth
// @Summary   		Delete Post
// @Description 	Api for delete a post
// @Tags 			post
// @Accept 			json
// @Produce 		json
// @Param 			id path string true "Post ID"
// @Success 		200 {object} bool
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/post/{id} [DELETE]
func (h *HandlerV1) DeletePost(c *gin.Context) {
	var (
		jspbMarshal protojson.MarshalOptions
	)
	jspbMarshal.UseProtoNames = true

	duration, err := time.ParseDuration(h.Config.Context.Timeout)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	userID := c.Param("id")

	err = h.Service.Post().DeletePost(ctx, &entity.DeleteReq{
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
// @Summary   		Get Post
// @Description 	Api for getting a post
// @Tags 			post
// @Accept 			json
// @Produce 		json
// @Param 			id path string true "Post ID"
// @Success 		200 {object} models.Post
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/post/{id} [GET]
func (h *HandlerV1) GetPost(c *gin.Context) {
	var (
		jspbMarshal protojson.MarshalOptions
	)
	jspbMarshal.UseProtoNames = true

	duration, err := time.ParseDuration(h.Config.Context.Timeout)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	userID := c.Param("id")

	filter := map[string]string{
		"id": userID,
	}
	post, err := h.Service.Post().GetPost(ctx, &entity.GetReq{
		Filter: filter,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, models.Post{
		Id:         userID,
		UserId:     post.UserId,
		Theme:      post.Theme,
		Path:       post.Path,
		Science:    post.Science,
		Views:      post.Views,
		CategoryId: post.CategoryId,
	})
}

// @Security  		BearerAuth
// @Summary   		Get  Delete Post
// @Description 	Api for getting a deleted post
// @Tags 			post
// @Accept 			json
// @Produce 		json
// @Param 			id path string true "Post ID"
// @Success 		200 {object} models.Post
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/del/post/{id} [GET]
func (h *HandlerV1) GetDelPost(c *gin.Context) {
	var (
		jspbMarshal protojson.MarshalOptions
	)
	jspbMarshal.UseProtoNames = true

	duration, err := time.ParseDuration(h.Config.Context.Timeout)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	userID := c.Param("id")

	filter := map[string]string{
		"id":  userID,
		"del": "",
	}
	post, err := h.Service.Post().GetPost(ctx, &entity.GetReq{
		Filter: filter,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, models.Post{
		Id:         userID,
		UserId:     post.UserId,
		Theme:      post.Theme,
		Path:       post.Path,
		Science:    post.Science,
		Views:      post.Views,
		CategoryId: post.CategoryId,
	})
}

// @Security  		BearerAuth
// @Summary   		List Post
// @Description 	Api for getting list post
// @Tags 			post
// @Accept 			json
// @Produce 		json
// @Param 			page query uint64 true "Page"
// @Param 			limit query uint64 true "Limit"
// @Success 		200 {object} models.ListPost
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/posts [GET]
func (h *HandlerV1) ListPost(c *gin.Context) {
	var (
		jspbMarshal protojson.MarshalOptions
	)
	jspbMarshal.UseProtoNames = true

	duration, err := time.ParseDuration(h.Config.Context.Timeout)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), duration)
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
	filter := map[string]string{}
	listPost, err := h.Service.Post().ListPost(ctx, &entity.ListReq{
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

	var posts []*models.Post
	for _, post := range listPost.Post {
		posts = append(posts, &models.Post{
			Id:         post.Id,
			UserId:     post.UserId,
			Theme:      post.Theme,
			Path:       post.Path,
			Views:      post.Views,
			CategoryId: post.CategoryId,
			Science:    post.Science,
		})
	}

	c.JSON(http.StatusOK, models.ListPost{
		Post:       posts,
		TotalCount: int(listPost.TotalCount),
	})
}

// @Security  		BearerAuth
// @Summary   		List Post
// @Description 	Api for getting user's posts
// @Tags 			post
// @Accept 			json
// @Produce 		json
// @Param 			page query int true "Page"
// @Param 			limit query int true "Limit"
// @Param 			id query string true "User Id"
// @Success 		200 {object} models.ListPost
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/user/posts [GET]
func (h *HandlerV1) GetAllPostByUserId(c *gin.Context) {
	var (
		body        models.GetAll
		jspbMarshal protojson.MarshalOptions
	)
	jspbMarshal.UseProtoNames = true

	duration, err := time.ParseDuration(h.Config.Context.Timeout)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	body.UserId = c.Query("id")
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
	body.Page = pageInt
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	body.Limit = limitInt

	offset := (body.Page - 1) * body.Limit

	filter := map[string]string{
		"user_id": body.UserId,
	}
	listPost, err := h.Service.Post().ListPost(ctx, &entity.ListReq{
		Offset: offset,
		Limit:  body.Limit,
		Filter: filter,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	var posts []*models.Post
	for _, post := range listPost.Post {
		posts = append(posts, &models.Post{
			Id:         post.Id,
			UserId:     post.UserId,
			Theme:      post.Theme,
			Path:       post.Path,
			Views:      post.Views,
			CategoryId: post.CategoryId,
			Science:    post.Science,
		})
	}

	c.JSON(http.StatusOK, models.ListPost{
		Post:       posts,
		TotalCount: int(listPost.TotalCount),
	})
}
