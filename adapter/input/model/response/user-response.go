package response

type UserErrorResponse struct {
	Error string `json:"error"`
}

type UserSuccessResponse struct {
	Message string `json:"message"`
}

type UserDataResponse struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Role  int    `json:"role"`
	Ativo bool   `json:"ativo"`
}

type GetUsersResponse struct {
	Users []UserDataResponse `json:"users"`
}
