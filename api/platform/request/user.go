package request

type RegisterRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,is_password_complex"`
	ConfirmPassword string `json:"confirm_password" validate:"required,empty_string,eqfield=Password"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,empty_string"`
}
