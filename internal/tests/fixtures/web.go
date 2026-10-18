package fixtures

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"

	"github.com/mathcale/go-api-boilerplate/internal/domain/user"
	"github.com/mathcale/go-api-boilerplate/internal/infra/web/handlers/dto"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/jwt"
	"github.com/mathcale/go-api-boilerplate/internal/usecases/auth"
)

const (
	WebName            = "Jane"
	WebSurname         = "Doe"
	WebEmail           = "jane@example.com"
	WebPassword        = "secret123"
	WebNewPassword     = "new-password123"
	WebCode            = "the-code"
	WebBadCode         = "bad-code"
	WebAccessToken     = "at"
	WebRefreshToken    = "rt"
	WebOldRefreshToken = "old-refresh-token"
	WebBadRefreshToken = "bad-token"
	WebInvalidUserID   = "not-a-uuid"
	WebShortUserID     = "11111111-1111-1111-1111-11111111111"
	WebMalformedBody   = "{"
	WebEmptyBody       = "{}"
)

func WebUserID() uuid.UUID {
	return uuid.New()
}

func WebTokenPair() *auth.TokenPair {
	return &auth.TokenPair{AccessToken: WebAccessToken, RefreshToken: WebRefreshToken}
}

func WebCreatedUser() *user.User {
	return &user.User{
		ID:      uuid.New(),
		Name:    WebName,
		Surname: WebSurname,
		Email:   WebEmail,
		Active:  false,
	}
}

func WebAuthenticatedUser(id uuid.UUID) *user.User {
	return &user.User{ID: id, Name: WebName, Email: WebEmail}
}

func WebSignInBody() string {
	return webJSON(dto.SignInInput{Email: WebEmail, Password: WebPassword})
}

func WebSignInMissingPasswordBody() string {
	return webJSON(map[string]string{"email": WebEmail})
}

func WebSignUpBody() string {
	return webJSON(dto.SignUpInput{
		Name:     WebName,
		Surname:  WebSurname,
		Email:    WebEmail,
		Password: WebPassword,
	})
}

func WebSignUpMissingFieldsBody() string {
	return webJSON(map[string]string{"name": WebName})
}

func WebRefreshTokenBody(token string) string {
	return webJSON(dto.RefreshTokenInput{RefreshToken: token})
}

func WebUserCodeBody(userID, code string) string {
	return webJSON(dto.ConfirmAccountInput{UserID: userID, Code: code})
}

func WebEmailBody() string {
	return webJSON(dto.SetRecoveryCodeInput{Email: WebEmail})
}

func WebUpdatePasswordBody(userID, code string) string {
	return webJSON(dto.UpdatePasswordInput{
		UserID:      userID,
		Code:        code,
		NewPassword: WebNewPassword,
	})
}

func webJSON(v interface{}) string {
	out, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return string(out)
}

func WebVerifiedAccessToken(userID uuid.UUID, version int) *jwt.Token {
	return &jwt.Token{
		Subject:      userID.String(),
		ExtraClaims:  jwt.ExtraClaims{Roles: WebRoles()},
		TokenVersion: version,
	}
}

func WebAuthState(version int, active bool) *user.AuthState {
	return &user.AuthState{TokenVersion: version, Active: active}
}

func WebRoles() []string {
	return []string{"user", "manager"}
}

func WebOversizedBody(size int) string {
	return `{"email":"` + strings.Repeat("a", size) + `"}`
}

func WebBodyWithUnknownField() string {
	return `{"email":"` + WebEmail + `","password":"` + WebPassword + `","admin":true}`
}

func WebBodyWithTrailingData() string {
	return WebSignInBody() + `{"email":"other@example.com"}`
}

func WebBodyWithTrailingGarbage() string {
	return WebSignInBody() + ` garbage`
}

func WebValidSignUpInput() dto.SignUpInput {
	avatar := "https://example.com/a.png"

	return dto.SignUpInput{
		Name:      WebName,
		Surname:   WebSurname,
		AvatarURL: &avatar,
		Email:     WebEmail,
		Password:  WebPassword,
	}
}

func WebSignUpInputWithAvatar(avatar string) dto.SignUpInput {
	in := WebValidSignUpInput()
	in.AvatarURL = &avatar

	return in
}

func WebSignUpInputWithPassword(password string) dto.SignUpInput {
	in := WebValidSignUpInput()
	in.Password = password

	return in
}

func WebSignUpInputWithName(name string) dto.SignUpInput {
	in := WebValidSignUpInput()
	in.Name = name

	return in
}

func WebSignUpInputWithEmail(email string) dto.SignUpInput {
	in := WebValidSignUpInput()
	in.Email = email

	return in
}
