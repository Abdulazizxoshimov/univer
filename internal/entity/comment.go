package entity

import "time"

type Comment struct{
	Id string
	OwnerId string
	PostId string
	Message string
	Likes int
	Dislikes int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CommentUpdateReq struct {
	Id string
	Message string
	UpdatedAt time.Time
}
type CommentListRes struct{
	Comment []*Comment
	TotalCount int
}

type Like struct{
	CommentId string
	PostId string
	OwnerId string
	Status bool
}