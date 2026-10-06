package mappers

import (
	"github.com/lib/pq"

	"github.com/mathcale/go-api-boilerplate/internal/modules/user/domain"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/internal/models"
)

func UserModelToDomain(m models.User) domain.User {
	return domain.User{
		ID:        m.ID,
		Name:      m.Name,
		Surname:   m.Surname,
		AvatarURL: m.AvatarURL,
		Email:     m.Email,
		Password:  m.Password,
		Active:    m.Active,
		Roles:     []string(m.Roles),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func UserDomainToModel(u domain.User) models.User {
	roles := u.Roles
	if roles == nil {
		// pq.StringArray(nil).Value() marshals to SQL NULL, which violates the
		// roles NOT NULL constraint; a non-nil empty slice marshals to '{}'.
		roles = []string{}
	}

	return models.User{
		ID:        u.ID,
		Name:      u.Name,
		Surname:   u.Surname,
		AvatarURL: u.AvatarURL,
		Email:     u.Email,
		Password:  u.Password,
		Active:    u.Active,
		Roles:     pq.StringArray(roles),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
