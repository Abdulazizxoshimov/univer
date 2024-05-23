package models

import "mime/multipart"

type Post struct {
	Id         string
	UserId     string
	Theme      string
	Path       string
	Views      int
	Science    string
	CategoryId string
}

type PostCreate struct {
	Theme      string   `json:"theme"`
	Science    string   `json:"science"`
	CategoryId string   `json:"category_id"`
}

type File struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

type PostUpdateReq struct {
	Id         string
	Theme      string
	Science    string
	CategoryId string
}

type ListPost struct {
	Post       []*Post
	TotalCount int
}

type GetAll struct {
	Page  int
	Limit int
	UserId    string
}
