package mappers

import (
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/domain"
	"github.com/mathcale/go-api-boilerplate/internal/modules/user/internal/models"
)

func ConfirmationCodeModelToDomain(m models.ConfirmationCode) domain.ConfirmationCode {
	return domain.ConfirmationCode{
		ID:        m.ID,
		UserID:    m.UserID,
		Code:      m.Code,
		Purpose:   domain.Purpose(m.Purpose),
		ExpiresAt: m.ExpiresAt,
		Used:      m.Used,
		CreatedAt: m.CreatedAt,
	}
}

func ConfirmationCodeDomainToModel(c domain.ConfirmationCode) models.ConfirmationCode {
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
