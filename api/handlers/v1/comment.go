package v1

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"
	"univer/api/models"
	"univer/internal/entity"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
)

// @Security      BearerAuth
// @Summary  	  Create Comment
// @Description   This api for create commment to post
// @Tags   		  comment
// @Accept 	      json
// @Produce 	  json
// @Param 		  comment body models.CommentCreate true "Comment Create Model"
// @Succes        201  {object} models.CreateResponse
// @Failure       401 {object} models.Error
// @Failure       403 {object} models.Error
// @Failure       500 {object} models.Error
// @Router        /v1/comment  [POST]
func (h *HandlerV1) CreateComment(c *gin.Context) {
	var (
		body        models.CommentCreate
		jspbMarshal protojson.MarshalOptions
	)
	jspbMarshal.UseProtoNames = true

	duration, err := time.ParseDuration(h.Config.Context.Timeout)
	if err != nil {
		c.JSON(500, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	err = c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(500, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	userId, statusCode := GetIdFromToken(c.Request, &h.Config)
	if statusCode != 0 {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "oops something went wrong",
		})
	}
	Comment, err := h.Service.Comment().CreateComment(ctx, &entity.Comment{
		OwnerId: userId,
		PostId:  body.PostId,
		Message: body.Message,
	})
	if err != nil {
		c.JSON(401, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	c.JSON(http.StatusCreated, models.CreateResponse{
		Id: Comment.Id,
	})
}

// @Security  		BearerAuth
// @Summary   		Update Comment
// @Description 	Api for update a Comment
// @Tags 			comment
// @Accept 			json
// @Produce 		json
// @Param 			comment body models.CommentUpdate true "Update Comment Model"
// @Success 		200 {object} models.CommentUpdate
// @Failure 		400 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/comment [PUT]
func (h *HandlerV1) UpdateComment(c *gin.Context) {
	var (
		body        models.CommentUpdate
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

	comment, err := h.Service.Comment().UpdateComment(ctx, &entity.CommentUpdateReq{
		Id:      body.Id,
		Message: body.Message,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, models.CommentUpdate{
		Id:      comment.Id,
		Message: comment.Message,
	})
}

// @Security  		BearerAuth
// @Summary   		Delete Comment
// @Description 	Api for delete a comment
// @Tags 			comment
// @Accept 			json
// @Produce 		json
// @Param 			id path string true "Comment ID"
// @Success 		200 {object} bool
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/comment/{id} [DELETE]
func (h *HandlerV1) DeleteComment(c *gin.Context) {
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

	err = h.Service.Comment().DeleteComment(ctx, &entity.DeleteReq{
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
// @Summary   		Get Comment
// @Description 	Api for getting a comment
// @Tags 			comment
// @Accept 			json
// @Produce 		json
// @Param 			id path string true "Comment ID"
// @Success 		200 {object} models.Comment
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/comment/{id} [GET]
func (h *HandlerV1) GetComment(c *gin.Context) {
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
	comment, err := h.Service.Comment().GetComment(ctx, &entity.GetReq{
		Filter: filter,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, models.Comment{
		Id:       userID,
		OwnerId:  comment.OwnerId,
		PostId:   comment.PostId,
		Message:  comment.Message,
		Likes:    comment.Likes,
		Dislikes: comment.Dislikes,
	})
}

// @Security  		BearerAuth
// @Summary   		List Comment
// @Description 	Api for getting list comment
// @Tags 			comment
// @Accept 			json
// @Produce 		json
// @Param 			page query uint64 true "Page"
// @Param 			limit query uint64 true "Limit"
// @Success 		200 {object} models.ListComment
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/comments [GET]
func (h *HandlerV1) ListComment(c *gin.Context) {
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
	listComment, err := h.Service.Comment().ListComment(ctx, &entity.ListReq{
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

	var comments []*models.Comment
	for _, comment := range listComment.Comment {
		comments = append(comments, &models.Comment{
			Id:       comment.Id,
			PostId:   comment.PostId,
			OwnerId:  comment.OwnerId,
			Message:  comment.Message,
			Likes:    comment.Likes,
			Dislikes: comment.Dislikes,
		})
	}

	c.JSON(http.StatusOK, models.ListComment{
		Comment:    comments,
		TotalCount: int(listComment.TotalCount),
	})
}

// @Security  		BearerAuth
// @Summary   		List Comment
// @Description 	Api for getting user's comment
// @Tags 			comment
// @Accept 			json
// @Produce 		json
// @Param 			page query int true "Page"
// @Param 			limit query int true "Limit"
// @Param 			id query string true "User Id"
// @Success 		200 {object} models.ListComment
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/user/comments [GET]
func (h *HandlerV1) GetAllCommentByUserId(c *gin.Context) {
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
		"owner_id": body.UserId,
	}
	listComment, err := h.Service.Comment().ListComment(ctx, &entity.ListReq{
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

	var comments []*models.Comment
	for _, comment := range listComment.Comment {
		comments = append(comments, &models.Comment{
			Id:       comment.Id,
			OwnerId:  comment.OwnerId,
			PostId:   comment.PostId,
			Message:  comment.Message,
			Likes:    comment.Likes,
			Dislikes: comment.Dislikes,
		})
	}

	c.JSON(http.StatusOK, models.ListComment{
		Comment:    comments,
		TotalCount: int(listComment.TotalCount),
	})
}

// @Security  		BearerAuth
// @Summary   		List Comment
// @Description 	Api for getting post's comment
// @Tags 			comment
// @Accept 			json
// @Produce 		json
// @Param 			page query int true "Page"
// @Param 			limit query int true "Limit"
// @Param 			id query string true "User Id"
// @Success 		200 {object} models.ListComment
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/post/comments [GET]
func (h *HandlerV1) GetAllCommentByPostId(c *gin.Context) {
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
		"post_id": body.UserId,
	}
	listComment, err := h.Service.Comment().ListComment(ctx, &entity.ListReq{
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

	var comments []*models.Comment
	for _, comment := range listComment.Comment {
		comments = append(comments, &models.Comment{
			Id:       comment.Id,
			OwnerId:  comment.OwnerId,
			PostId:   comment.PostId,
			Message:  comment.Message,
			Likes:    comment.Likes,
			Dislikes: comment.Dislikes,
		})
	}

	c.JSON(http.StatusOK, models.ListComment{
		Comment:    comments,
		TotalCount: int(listComment.TotalCount),
	})
}

// @Security        BearerAuth
// @Summary         Create Like
// @Description     This api for create coment's like
// @Tags            comment
// @Accept          json
// @Produce         json
// @Param           like body models.CreateLike true "Create Like Model"
// @Success 		201 {object} models.Like
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/comment/like [POST]
func (h *HandlerV1) CreateLike(c *gin.Context) {
	var (
		body        models.CreateLike
		jspbMarshal protojson.MarshalOptions
	)
	jspbMarshal.UseProtoNames = true

	duration, err := time.ParseDuration(h.Config.Context.Timeout)
	if err != nil {
		c.JSON(500, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	err = c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(500, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	userId, statusCode := GetIdFromToken(c.Request, &h.Config)
	if statusCode != 0 {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "oops something went wrong",
		})
	}
	status, err := h.Service.Comment().CreateLike(ctx, &entity.Like{
		OwnerId:   userId,
		PostId:    body.PostId,
		CommentId: body.CommentId,
		Status:    true,
	})
	if err != nil {
		c.JSON(401, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	c.JSON(http.StatusCreated, models.Like{
		CommentId: body.CommentId,
		OwnerId:   userId,
		PostId:    body.CommentId,
		Status:    status,
	})
}

// @Security        BearerAuth
// @Summary         Create DisLike
// @Description     This api for create coment's like
// @Tags            comment
// @Accept          json
// @Produce         json
// @Param           like body  models.CreateLike true "Create DisLike Model"
// @Success 		201 {object} models.Like
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/comment/dislike [POST]
func (h *HandlerV1) CreateDisLike(c *gin.Context) {
	var (
		body        models.CreateLike
		jspbMarshal protojson.MarshalOptions
	)
	jspbMarshal.UseProtoNames = true

	duration, err := time.ParseDuration(h.Config.Context.Timeout)
	if err != nil {
		c.JSON(500, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	err = c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(500, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	userId, statusCode := GetIdFromToken(c.Request, &h.Config)
	if statusCode != 0 {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: "oops something went wrong",
		})
	}
	status, err := h.Service.Comment().CreateDislike(ctx, &entity.Like{
		OwnerId:   userId,
		PostId:    body.PostId,
		CommentId: body.CommentId,
		Status:    true,
	})
	if err != nil {
		c.JSON(401, models.Error{
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	c.JSON(http.StatusCreated, models.Like{
		CommentId: body.CommentId,
		OwnerId:   userId,
		PostId:    body.CommentId,
		Status:    status,
	})
}
