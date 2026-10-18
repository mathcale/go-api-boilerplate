package fixtures

import (
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/infra/database/models"
	"github.com/mathcale/go-api-boilerplate/internal/infra/email"
)

const (
	InfraUserEmail      = "jane@example.com"
	InfraUnknownEmail   = "nobody@example.com"
	InfraConfirmPurpose = "account_confirmation"
)

func InfraDomainUser() user.User {
	return user.User{
		ID:    uuid.New(),
		Name:  "Jane",
		Email: InfraUserEmail,
	}
}

func InfraDomainConfirmationCode(userID uuid.UUID) user.ConfirmationCode {
	return user.NewConfirmationCode(
		userID,
		user.PurposeAccountConfirmation,
		user.AccountConfirmationTTL,
	)
}

func InfraDomainRecoveryCode(userID uuid.UUID) user.ConfirmationCode {
	return user.NewConfirmationCode(
		userID,
		user.PurposePasswordRecovery,
		user.PasswordRecoveryTTL,
	)
}

func InfraDomainRefreshToken() user.RefreshToken {
	return user.NewRefreshTokenFamily(uuid.New(), uuid.New(), time.Now().Add(time.Hour))
}

func InfraModelUser() models.User {
	return models.User{
		ID:    uuid.New(),
		Email: InfraUserEmail,
	}
}

func InfraModelConfirmationCode(userID uuid.UUID) models.ConfirmationCode {
	return models.ConfirmationCode{
		ID:      uuid.New(),
		UserID:  userID,
		Purpose: InfraConfirmPurpose,
	}
}

func InfraModelRefreshToken() models.RefreshToken {
	return models.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		FamilyID:  uuid.New(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
}

func InfraUserRows(id uuid.UUID) *sqlmock.Rows {
	columns := []string{
		"id", "name", "surname", "avatar_url", "email",
		"password", "active", "roles", "created_at", "updated_at", "token_version",
	}

	return sqlmock.NewRows(columns).AddRow(
		id, "Jane", "Doe", nil, InfraUserEmail, "hashed", true, "{}", time.Now(), time.Now(), 4,
	)
}

func InfraUserAuthStateRows(version int, active bool) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"token_version", "active"}).AddRow(version, active)
}

func InfraModelUserAuthState() models.UserAuthState {
	return models.UserAuthState{TokenVersion: 4, Active: true}
}

func InfraConfirmationCodeRows(id, userID uuid.UUID, code string) *sqlmock.Rows {
	columns := []string{"id", "user_id", "code", "purpose", "expires_at", "used", "created_at"}

	return sqlmock.NewRows(columns).AddRow(
		id, userID, code, InfraConfirmPurpose, time.Now(), false, time.Now(),
	)
}

func InfraRefreshTokenRows(id, userID, familyID uuid.UUID) *sqlmock.Rows {
	columns := []string{
		"id", "user_id", "family_id", "expires_at", "used_at", "revoked_at", "created_at",
	}

	return sqlmock.NewRows(columns).AddRow(
		id, userID, familyID, time.Now().Add(24*time.Hour), nil, nil, time.Now(),
	)
}

func InfraEmailSendInput() email.SendInput {
	return email.SendInput{
		RecipientName:  "Jane",
		RecipientEmail: InfraUserEmail,
		Subject:        "Hello",
		HTMLBody:       "<p>Hello Jane</p>",
	}
}

func InfraSMTPSendInputWithName(name string) email.SendInput {
	in := InfraEmailSendInput()
	in.RecipientName = name

	return in
}
