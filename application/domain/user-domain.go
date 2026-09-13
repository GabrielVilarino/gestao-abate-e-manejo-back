package domain

const (
	RoleAdmin = 1
	RoleUser  = 2
)

type User struct {
	ID       int
	Nome     string
	Password string
	Email    string
	Role     int
	Ativo    bool
}
