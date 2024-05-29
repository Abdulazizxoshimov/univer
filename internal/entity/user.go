package entity

import "time"

type User struct {
	Id           string
	UserName     string
	Email        string
	PhoneNumber  string
	Password     string
	Bio          string
	ImageUrl     string
	RefreshToken string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UpdateRefresh struct {
	UserID       string
	RefreshToken string
}

type UpdatePassword struct {
	UserID      string
	NewPassword string
}

type IsUnique struct {
	Email string
}

type Response struct {
	Status bool
}
type ListUserRes struct {
	User       []*User
	TotalCount int64
}
type ListReq struct{
   Limit int
   Offset int
   Filter map[string]string
}
type DeleteReq struct{
	Id string
	DeletedAt time.Time	
}

type GetReq struct {
	Filter map[string]string
}
type UpdateProfile struct {
	Id string
	ImageUrl  string
}
