package bcrypt

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
)

type (
	Password interface {
		Hash(password string) (*string, error)
		Verify(password string, hashedPassword string) error
	}

	password struct {
		cost int
	}
)

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
			apperror.PackageOrigin, "bcrypt", nil, nil,
		)
	}

	hashStr := string(hash)

	return &hashStr, nil
}

func (b *password) Verify(password string, hashedPassword string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		be := apperror.BE1010_INVALID_CREDENTIALS

		return apperror.New(
			err, "invalid credentials",
			apperror.ValidationKind, apperror.PackageOrigin, "bcrypt", &be, nil,
		)
	}

	return nil
}
