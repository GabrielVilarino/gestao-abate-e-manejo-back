package domain

type User struct {
	ID       int
	Nome     string
	Password string
	Email    string
	Role     int
	Ativo    bool
}
