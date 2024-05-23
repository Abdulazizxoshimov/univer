package models

type (
	Login struct {
		UserNameOrEmail string `json:"usernameoremail" example:"abdulazizxoshimov22@gmail.com"`
		Password        string `json:"password" example:"@Abdulaziz2004"`
	}

	Otp struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}

	ResetPassword struct {
		Otp         string `json:"otp"`
		Email       string `json:"email"`
		NewPassword string `json:"new_password"`
	}



	TokenResp struct {
		ID      string `json:"user_id"`
		Access  string `json:"access_token"`
		Refresh string `json:"refresh_token"`
		Role    string `json:"role"`
	}
    
)
