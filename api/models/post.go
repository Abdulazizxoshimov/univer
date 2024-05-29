package models

import "mime/multipart"

type Post struct {
	Id          string
	UserId      string
	Theme       string
	Path        string
	Views       int
	Science     string
	CategoryId  string
	Price       float64
	PriceStatus bool
}

type PostCreate struct {
	Theme      string  `json:"theme"`
	Science    string  `json:"science"`
	CategoryId string  `json:"category_id"`
	Price      float64 `json:"price"`
}

type File struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

type PostUpdateReq struct {
	Id         string    `json:"id" binding:"required"`
	Theme      string    `json:"theme" binding:"required"`
	Science    string    `json:"science" binding:"required"`
	CategoryId string    `json:"category_id" binding:"required"`
	Price      float64   `json:"price,omitempty"`
}


type ListPost struct {
	Post       []*Post
	TotalCount int
}

type GetAll struct {
	Page   int
	Limit  int
	UserId string
}
