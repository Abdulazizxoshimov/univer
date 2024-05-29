package v1

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"univer/api/models"
	"univer/internal/entity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"google.golang.org/protobuf/encoding/protojson"
)

// @Security      BearerAuth
// @Summary       Create Post
// @Description   Api for create a new Post
// @Tags          post
// @Accept        multipart/form-data
// @Produce       json
// @Param         theme query string true "Theme"
// @Param         science query string true "Science"
// @Param         id query string true "Category Id"
// @Param         price query string false "Price"
// @Param         file formData file true "File"
// @Success       201 {object} models.CreateResponse
// @Failure       400 {object} models.Error
// @Failure       401 {object} models.Error
// @Failure       403 {object} models.Error
// @Failure       500 {object} models.Error
// @Router        /v1/post [POST]
func (h *HandlerV1) CreatePost(c *gin.Context) {
	var (
		file        models.File
		body        models.PostCreate
		jspbMarshal protojson.MarshalOptions
	)
	jspbMarshal.UseProtoNames = true

	ctx, cancel := context.WithTimeout(context.Background(), h.Config.Context.Timeout)
	defer cancel()

	body.CategoryId = c.Query("id")
	body.Science = c.Query("science")
	body.Theme = c.Query("theme")
	price := c.Query("price")
	var err error
	body.Price, err = strconv.ParseFloat(price, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

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
		".zip":  true,
	}

	if !allowedExtensions[ext] {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "Only .pdf, .doc, .docx, .ppt, and .pptx format files are accepted"})
		return
	}

	id := uuid.New().String()
	objectName := id + ext

	fileHeader, err := file.File.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}
	defer fileHeader.Close()

	contentType := c.ContentType()
	_, err = h.MinIO.PutObject(ctx, h.Config.Minio.FileUploadBucketName, objectName, fileHeader, file.File.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err)
		return
	}

	minioURL := fmt.Sprintf("http://%s/%s/%s", h.Config.Minio.Endpoint, h.Config.Minio.FileUploadBucketName, objectName)

	userId, statusCode := GetIdFromToken(c.Request, &h.Config)
	if statusCode != 0 {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "oops something went wrong",
		})
	}
	role, statusCode := GetRoleFromToken(c.Request, &h.Config)
	if statusCode != 0 {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "oops something went wrong",
		})
	}

	if body.Price > 0 && role == "prouser" {
		post, err := h.Service.Post().CreatePost(ctx, &entity.Post{
			Id:          id,
			UserId:      userId,
			Theme:       body.Theme,
			Path:        minioURL,
			Science:     body.Science,
			CategoryId:  body.CategoryId,
			PriceStatus: true,
			Price:       body.Price,
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
	} else {
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
	role, statusCode := GetRoleFromToken(c.Request, &h.Config)
	if statusCode != 0 {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "oops something went wrong",
		})
	}

	if role == "prouser" {
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
	} else {
		post, err := h.Service.Post().UpdatePost(ctx, &entity.PostUpdateReq{
			Id:         body.Id,
			Theme:      body.Theme,
			Science:    body.Science,
			CategoryId: body.CategoryId,
			Price:      body.Price,
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

	ctx, cancel := context.WithTimeout(context.Background(), h.Config.Context.Timeout)
	defer cancel()

	userID := c.Param("id")

	err := h.Service.Post().DeletePost(ctx, &entity.DeleteReq{
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

	ctx, cancel := context.WithTimeout(context.Background(), h.Config.Context.Timeout)
	defer cancel()

	id := c.Param("id")

	userId, statusCode := GetIdFromToken(c.Request, &h.Config)
	if statusCode != 0 {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "oops something went wrong",
		})
	}

	filter := map[string]string{
		"id":      id,
		"user_id": userId,
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
		Id:          id,
		UserId:      post.UserId,
		Theme:       post.Theme,
		Path:        post.Path,
		Science:     post.Science,
		Views:       post.Views,
		CategoryId:  post.CategoryId,
		PriceStatus: post.PriceStatus,
		Price:       post.Price,
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

	ctx, cancel := context.WithTimeout(context.Background(), h.Config.Context.Timeout)
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
		Id:          userID,
		UserId:      post.UserId,
		Theme:       post.Theme,
		Path:        post.Path,
		Science:     post.Science,
		Views:       post.Views,
		CategoryId:  post.CategoryId,
		PriceStatus: post.PriceStatus,
		Price:       post.Price,
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
			Id:          post.Id,
			UserId:      post.UserId,
			Theme:       post.Theme,
			Path:        post.Path,
			Views:       post.Views,
			CategoryId:  post.CategoryId,
			Science:     post.Science,
			Price:       post.Price,
			PriceStatus: post.PriceStatus,
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

	ctx, cancel := context.WithTimeout(context.Background(), h.Config.Context.Timeout)
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
			Id:          post.Id,
			UserId:      post.UserId,
			Theme:       post.Theme,
			Path:        post.Path,
			Views:       post.Views,
			CategoryId:  post.CategoryId,
			Science:     post.Science,
			Price:       post.Price,
			PriceStatus: post.PriceStatus,
		})
	}

	c.JSON(http.StatusOK, models.ListPost{
		Post:       posts,
		TotalCount: int(listPost.TotalCount),
	})
}
