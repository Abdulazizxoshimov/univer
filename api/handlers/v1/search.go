package v1

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"univer/api/models"
	"univer/internal/entity"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
)

// @Security  		BearerAuth
// @Summary   		Search
// @Description 	Api for searching by theme
// @Tags 			search
// @Accept 			json
// @Produce 		json
// @Param 			page query int true "Page"
// @Param 			limit query int true "Limit"
// @Param 			theme query string true "Theme"
// @Param           science  query string false "Science"
// @Param           priceStatus  query bool false "Price Satatus"
// @Param           category  query string false "Category" Enums(slayt, mustaqil ish,kurs ishi)
// @Success 		200 {object} models.ListPost
// @Failure 		404 {object} models.Error
// @Failure 		401 {object} models.Error
// @Failure 		403 {object} models.Error
// @Failure 		500 {object} models.Error
// @Router 			/v1/search [GET]
func (h *HandlerV1) Search(c *gin.Context) {
	var (
		jspbMarshal protojson.MarshalOptions
	)
	jspbMarshal.UseProtoNames = true

	ctx, cancel := context.WithTimeout(context.Background(), h.Config.Context.Timeout)
	defer cancel()
	theme := c.Query("theme")
	page := c.Query("page")
	limit := c.Query("limit")
	science := c.Query("science")
	category := c.Query("category")
	priceStatus := c.Query("priceStatus")
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
		"theme": theme,
	}
	if science != "" {
		filter["science"] = science
	}
	if category != "" {
		filter["category_id"] = category
	}
	if priceStatus != "" {
		filter["price_status"] = priceStatus
	}
	listPost, err := h.Service.Post().Search(ctx, &entity.ListReq{
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
