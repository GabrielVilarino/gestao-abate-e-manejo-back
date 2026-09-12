package security

import "golang.org/x/crypto/bcrypt"

type HashPort struct{}

func NewHashPort() *HashPort {
	return &HashPort{}
}

func (h *HashPort) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (h *HashPort) Compare(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(password),
	)
}
