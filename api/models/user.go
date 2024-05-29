package models

type UserRegister struct {
	UserName    string
	Email       string
	Password    string
}

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
}

type UpdateReq struct {
	Id          string
	UserName    string
	Email       string
	PhoneNumber string
	Bio         string
}

type Response struct {
	Response string
}

type UserResponse struct {
	Id          string `json:"id"`
	UserName    string `json:"username"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Bio         string `json:"bio"`
	ImageUrl    string `json:"image_url"`
	Role        string `json:"role"`
	Refresh     string `json:"refresh_token"`
	Access      string `json:"access_token"`
}


type CreateResponse struct {
	Id string `json:"id"`
}

type ListUser struct {
	User  []UserResponse `json:"user"`
	Total uint64 `json:"totcal_count"`
}

type UpdatePasswordReq struct {
	Id       string `json:"id"`
	Password string `json:"password"`
}

type GoogleUser struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	PictureUrl    string `json:"picture"`
	Locale        string `json:"locale"`
}
