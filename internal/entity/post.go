package entity

import "time"

type Post struct {
	Id         string
	UserId     string
	Theme      string
	Path       string
	Views      int
	Science    string
	CategoryId string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PostUpdateReq struct{
	Id         string
	Theme      string
	Path       string
	Science    string
	CategoryId string
	UpdatedAt  time.Time
}

type PostListRes struct {
	Post []*Post
	TotalCount int
}
type Search struct {
	Theme string
}
