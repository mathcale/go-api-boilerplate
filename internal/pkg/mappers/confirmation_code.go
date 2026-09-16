package mappers

import (
	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/infra/database/models"
)

func ConfirmationCodeModelToDomain(m models.ConfirmationCode) user.ConfirmationCode {
	return user.ConfirmationCode{
		ID:        m.ID,
		UserID:    m.UserID,
		Code:      m.Code,
		Purpose:   user.Purpose(m.Purpose),
		ExpiresAt: m.ExpiresAt,
		Used:      m.Used,
		CreatedAt: m.CreatedAt,
	}
}

func ConfirmationCodeDomainToModel(c user.ConfirmationCode) models.ConfirmationCode {
	return models.ConfirmationCode{
		ID:        c.ID,
		UserID:    c.UserID,
		Code:      c.Code,
		Purpose:   string(c.Purpose),
		ExpiresAt: c.ExpiresAt,
		Used:      c.Used,
		CreatedAt: c.CreatedAt,
	}
}
