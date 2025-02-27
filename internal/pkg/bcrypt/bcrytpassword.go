package bcrypt

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
)

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
		return nil, apperror.New(
			err, "password hashing failed", apperror.DependencyKind,
			apperror.PackageOrigin, "bcryptpassword", nil, nil,
		)
	}

	hashStr := string(hash)

	return &hashStr, nil
}

func (b *password) Verify(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
