package requests

type Register struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required"`
}

type Login struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RFToken struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
