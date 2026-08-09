package auth

type LoginCommand struct {
	Email    string `validate:"required,email,max=254"`
	Password string `validate:"required,min=6,max=128"`
	AppID    int    `validate:"required,gt=0"`
}

type RegisterCommand struct {
	Email    string  `validate:"required,email,max=254"`
	Password string  `validate:"required,min=6,max=128"`
	Username string  `validate:"omitempty,min=3,max=50,alphanum"`
	Roles    []int64 `validate:"min=0,max=10"`
}
