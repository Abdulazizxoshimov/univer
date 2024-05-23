package models


type Comment struct{
	Id string
	OwnerId string
	PostId string
	Message string
	Likes int
	Dislikes int
}

type CommentCreate struct{
	PostId string
	Message string
}

type CommentUpdate struct{
	Id string
	Message string
}

type  ListComment struct{
	Comment []*Comment
	TotalCount int
}

type Like struct{
	CommentId string
	PostId string
	OwnerId string
	Status bool
}

type CreateLike struct{
	CommentId string
	PostId string
	Status bool
}

type Metadata struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
