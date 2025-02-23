package bcrypt

import "golang.org/x/crypto/bcrypt"

type Password interface {
	Hash(password string) (*string, error)
	Verify(password string, hashedPassword string) error
}

type password struct {
	cost int
}

func NewPassword() Password {
	return &password{
		cost: bcrypt.DefaultCost,
	}
}

func (b *password) Hash(password string) (*string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return nil, err
	}

	hashStr := string(hash)

	return &hashStr, nil
}

func (b *password) Verify(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
