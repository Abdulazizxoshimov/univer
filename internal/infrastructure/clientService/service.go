package clientService

import "univer/internal/usecase"


type ServiceClient interface {
	User() usecase.User
	Category() usecase.Category
	Comment()   usecase.Comment
	Post() usecase.Post
}

type serviceClient struct{
	user usecase.User
	post usecase.Post
	comment usecase.Comment
	category usecase.Category
}

func New(user usecase.User, post usecase.Post, comment usecase.Comment, category usecase.Category)ServiceClient{
	return &serviceClient{
		user: user,
		post: post,
		category: category,
		comment: comment,
	}
}

func (s *serviceClient)User() usecase.User{
	return s.user
}
func (s *serviceClient)Category() usecase.Category{
	return s.category
}
func (s *serviceClient)Comment() usecase.Comment{
	return s.comment
}
func (s *serviceClient)Post() usecase.Post{
	return s.post
}
