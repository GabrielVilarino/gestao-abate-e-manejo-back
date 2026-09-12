package request

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserCreateRequest struct {
	Nome     string `json:"nome" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     int    `json:"role" binding:"required"`
}

type UserUpdateRequest struct {
	ID    int    `json:"id" binding:"required"`
	Nome  string `json:"nome" binding:"required"`
	Email string `json:"email" binding:"required"`
	Role  int    `json:"role" binding:"required"`
}
