package user

import (
	"context"
	"fmt"
	"html"

	"github.com/google/uuid"

	"go-api-boilerplate/internal/domain/gateway"
	"go-api-boilerplate/internal/domain/user"
	"go-api-boilerplate/internal/infra/email"
	"go-api-boilerplate/internal/pkg/mappers"
)

type userGateway struct {
	repo                       Repository
	emailClient                email.Client
	accountConfirmationBaseURL string
	passwordRecoveryBaseURL    string
}

func NewGateway(
	repo Repository,
	emailClient email.Client,
	accountConfirmationBaseURL string,
	passwordRecoveryBaseURL string,
) gateway.User {
	return &userGateway{
		repo:                       repo,
		emailClient:                emailClient,
		accountConfirmationBaseURL: accountConfirmationBaseURL,
		passwordRecoveryBaseURL:    passwordRecoveryBaseURL,
	}
}

func (g *userGateway) UserExistsIncludingInactive(ctx context.Context, email string) (*bool, error) {
	return g.repo.ExistsIncludingInactive(ctx, email)
}

func (g *userGateway) SaveUser(ctx context.Context, u user.User, code user.ConfirmationCode) error {
	return g.repo.Save(
		ctx,
		mappers.UserDomainToModel(u),
		mappers.ConfirmationCodeDomainToModel(code),
	)
}

func (g *userGateway) GetUserByEmailIncludingInactive(
	ctx context.Context,
	email string,
) (*user.User, error) {
	m, err := g.repo.GetIncludingInactive(ctx, email)
	if err != nil {
		return nil, err
	}

	if m == nil {
		return nil, nil
	}

	u := mappers.UserModelToDomain(*m)

	return &u, nil
}

func (g *userGateway) GetUserByID(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	m, err := g.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if m == nil {
		return nil, nil
	}

	u := mappers.UserModelToDomain(*m)

	return &u, nil
}

func (g *userGateway) GetUserAuthState(
	ctx context.Context,
	userID uuid.UUID,
) (*user.AuthState, error) {
	m, err := g.repo.GetAuthState(ctx, userID)
	if err != nil {
		return nil, err
	}

	if m == nil {
		return nil, nil
	}

	state := mappers.UserAuthStateModelToDomain(*m)

	return &state, nil
}

func (g *userGateway) UpdatePasswordWithCode(
	ctx context.Context,
	userID uuid.UUID,
	codeID uuid.UUID,
	hashedPassword string,
) (bool, error) {
	return g.repo.UpdatePasswordWithCode(ctx, userID, codeID, hashedPassword)
}

func (g *userGateway) ActivateUserWithCode(
	ctx context.Context,
	userID uuid.UUID,
	codeID uuid.UUID,
) (bool, error) {
	return g.repo.ActivateUserWithCode(ctx, userID, codeID)
}

func (g *userGateway) SaveConfirmationCode(ctx context.Context, code user.ConfirmationCode) error {
	return g.repo.SaveConfirmationCode(ctx, mappers.ConfirmationCodeDomainToModel(code))
}

func (g *userGateway) GetValidConfirmationCode(
	ctx context.Context,
	userID uuid.UUID,
	code string,
	purpose user.Purpose,
) (*user.ConfirmationCode, error) {
	m, err := g.repo.GetConfirmationCode(ctx, userID, code, string(purpose))
	if err != nil {
		return nil, err
	}

	if m == nil {
		return nil, nil
	}

	c := mappers.ConfirmationCodeModelToDomain(*m)

	return &c, nil
}

func (g *userGateway) SendConfirmationEmail(
	ctx context.Context,
	u user.User,
	code user.ConfirmationCode,
) error {
	link := fmt.Sprintf("%s?userID=%s&code=%s", g.accountConfirmationBaseURL, u.ID, code.Code)

	return g.emailClient.Send(email.SendInput{
		RecipientName:  u.Name,
		RecipientEmail: u.Email,
		Subject:        "Confirm your account",
		HTMLBody: fmt.Sprintf(
			"<p>Hi %s,</p><p>Confirm your account by clicking "+
				"<a href=\"%s\">this link</a>.</p>",
			html.EscapeString(u.Name), html.EscapeString(link),
		),
	})
}

func (g *userGateway) SendRecoveryEmail(
	ctx context.Context,
	u user.User,
	code user.ConfirmationCode,
) error {
	link := fmt.Sprintf("%s/%s?code=%s", g.passwordRecoveryBaseURL, u.ID, code.Code)

	return g.emailClient.Send(email.SendInput{
		RecipientName:  u.Name,
		RecipientEmail: u.Email,
		Subject:        "Reset your password",
		HTMLBody: fmt.Sprintf(
			"<p>Hi %s,</p><p>Reset your password by clicking "+
				"<a href=\"%s\">this link</a>.</p>",
			html.EscapeString(u.Name), html.EscapeString(link),
		),
	})
}

func (g *userGateway) SaveRefreshToken(ctx context.Context, t user.RefreshToken) error {
	return g.repo.SaveRefreshToken(ctx, mappers.RefreshTokenDomainToModel(t))
}

func (g *userGateway) GetRefreshToken(
	ctx context.Context,
	tokenID uuid.UUID,
) (*user.RefreshToken, error) {
	m, err := g.repo.GetRefreshToken(ctx, tokenID)
	if err != nil {
		return nil, err
	}

	if m == nil {
		return nil, nil
	}

	t := mappers.RefreshTokenModelToDomain(*m)

	return &t, nil
}

func (g *userGateway) MarkRefreshTokenUsed(
	ctx context.Context,
	tokenID uuid.UUID,
) (bool, error) {
	return g.repo.MarkRefreshTokenUsed(ctx, tokenID)
}

func (g *userGateway) RevokeRefreshTokenFamily(ctx context.Context, familyID uuid.UUID) error {
	return g.repo.RevokeRefreshTokenFamily(ctx, familyID)
}

func (g *userGateway) RevokeUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	return g.repo.RevokeUserRefreshTokens(ctx, userID)
}
