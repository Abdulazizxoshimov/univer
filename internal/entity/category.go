package entity

import "time"

type Category struct {
	Id        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
type UpdateCategory struct {
	Id        string
	Name      string
	UpdatedAt time.Time
}

type ListCategoryRes struct{
	Category []*Category
	Totalcount int
}
