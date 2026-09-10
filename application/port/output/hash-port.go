package output

type HashPort interface {
	Hash(password string) (string, error)
	Compare(password, hashedPassword string) error
}
