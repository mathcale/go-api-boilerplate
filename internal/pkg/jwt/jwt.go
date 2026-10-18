package jwt

import (
	"errors"
	"math"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	claimType         = "typ"
	claimTokenVersion = "tv"
)

type (
	ExtraClaims struct {
		Roles []string `json:"roles,omitempty"`
	}

	IssueTokenParams struct {
		UserID       string
		ExtraClaims  ExtraClaims
		TokenID      string
		TokenVersion int
	}

	JWTAuth interface {
		IssueAccessToken(params IssueTokenParams) (*Token, error)
		IssueRefreshToken(params IssueTokenParams) (*Token, error)
		VerifyAccessToken(at string) (*Token, error)
		VerifyRefreshToken(rt string) (*Token, error)
	}

	jwtAuth struct {
		logger          logger.Logger
		accessSecret    []byte
		accessLifetime  int
		refreshSecret   []byte
		refreshLifetime int
		issuer          string
		audience        string
		signingMethod   jwtlib.SigningMethod
	}

	Token struct {
		Token       *string     `json:"token,omitempty"`
		Issuer      string      `json:"iss,omitempty"`
		Subject     string      `json:"sub,omitempty"`
		ExtraClaims ExtraClaims `json:"extra_claims,omitempty"`
		Audience    []string    `json:"aud,omitempty"`
		ExpiresAt   *time.Time  `json:"exp,omitempty"`
		IssuedAt    *time.Time  `json:"iat,omitempty"`
		ID          string      `json:"jti,omitempty"`

		Type         string `json:"typ,omitempty"`
		TokenVersion int    `json:"tv"`
	}
)

func NewJWTAuth(
	l logger.Logger,
	secret, refreshSecret []byte,
	lifetime, refreshLifetime int,
	issuer, audience string,
) JWTAuth {
	return &jwtAuth{
		logger:          l,
		accessSecret:    secret,
		accessLifetime:  lifetime,
		refreshSecret:   refreshSecret,
		refreshLifetime: refreshLifetime,
		issuer:          issuer,
		audience:        audience,
		signingMethod:   jwtlib.SigningMethodHS512,
	}
}

func (j *jwtAuth) IssueAccessToken(params IssueTokenParams) (*Token, error) {
	return j.issue(params, TokenTypeAccess, j.accessSecret, j.accessLifetime)
}

func (j *jwtAuth) VerifyAccessToken(at string) (*Token, error) {
	return j.verify(at, TokenTypeAccess, j.accessSecret)
}

func (j *jwtAuth) IssueRefreshToken(params IssueTokenParams) (*Token, error) {
	return j.issue(params, TokenTypeRefresh, j.refreshSecret, j.refreshLifetime)
}

func (j *jwtAuth) VerifyRefreshToken(rt string) (*Token, error) {
	return j.verify(rt, TokenTypeRefresh, j.refreshSecret)
}

func (j *jwtAuth) issue(
	params IssueTokenParams,
	tokenType string,
	secret []byte,
	lifetime int,
) (*Token, error) {
	roles := params.ExtraClaims.Roles
	if roles == nil {
		roles = []string{}
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(lifetime) * time.Minute)

	claims := jwtlib.MapClaims{
		"sub": params.UserID,
		"extra_claims": map[string]interface{}{
			"roles": roles,
		},
		"iss":             j.issuer,
		"aud":             j.audience,
		"exp":             expiresAt.Unix(),
		"iat":             now.Unix(),
		claimType:         tokenType,
		claimTokenVersion: params.TokenVersion,
	}

	if params.TokenID != "" {
		claims["jti"] = params.TokenID
	}

	token := jwtlib.NewWithClaims(j.signingMethod, claims)

	j.logger.Debug("Token claims added", "sub", params.UserID)

	tokenStr, err := token.SignedString(secret)
	if err != nil {
		return nil, apperror.New(
			err, "token signing failed", apperror.DependencyKind,
			apperror.PackageOrigin, "jwt", nil, nil,
		)
	}

	return &Token{
		Token:        &tokenStr,
		Issuer:       j.issuer,
		Subject:      params.UserID,
		ExtraClaims:  ExtraClaims{Roles: roles},
		Audience:     []string{j.audience},
		ExpiresAt:    &expiresAt,
		IssuedAt:     &now,
		ID:           params.TokenID,
		Type:         tokenType,
		TokenVersion: params.TokenVersion,
	}, nil
}

func (j *jwtAuth) verify(token, expectedType string, secret []byte) (*Token, error) {
	t, err := jwtlib.Parse(token, func(token *jwtlib.Token) (interface{}, error) {
		return secret, nil
	},
		jwtlib.WithValidMethods([]string{j.signingMethod.Alg()}),
		jwtlib.WithIssuer(j.issuer),
		jwtlib.WithAudience(j.audience),
		jwtlib.WithExpirationRequired(),
	)
	if err != nil {
		return nil, apperror.New(
			err, "token parsing failed", apperror.ParseKind,
			apperror.PackageOrigin, "jwt", nil, nil,
		)
	}

	claims, ok := t.Claims.(jwtlib.MapClaims)
	if !ok || !t.Valid {
		return nil, invalidClaim("invalid_token", "invalid jwt token")
	}

	parsed, err := j.toToken(&token, claims)
	if err != nil {
		return nil, err
	}

	if parsed.Type != expectedType {
		return nil, invalidClaim("invalid_token_type", "unexpected jwt token type")
	}

	return parsed, nil
}

func (j *jwtAuth) toToken(token *string, claims jwtlib.MapClaims) (*Token, error) {
	iat, err := claims.GetIssuedAt()
	if err != nil || iat == nil {
		return nil, invalidClaim("invalid_iat", "iat claim parse failed")
	}

	exp, err := claims.GetExpirationTime()
	if err != nil || exp == nil {
		return nil, invalidClaim("invalid_exp", "exp claim parse failed")
	}

	issuer, err := claims.GetIssuer()
	if err != nil || issuer == "" {
		return nil, invalidClaim("invalid_iss", "iss claim parse failed")
	}

	subject, err := claims.GetSubject()
	if err != nil || subject == "" {
		return nil, invalidClaim("invalid_sub", "sub claim parse failed")
	}

	audience, err := claims.GetAudience()
	if err != nil || len(audience) == 0 {
		return nil, invalidClaim("invalid_aud", "aud claim parse failed")
	}

	tokenType, ok := claims[claimType].(string)
	if !ok || tokenType == "" {
		return nil, invalidClaim("invalid_typ", "typ claim parse failed")
	}

	version, ok := parseTokenVersion(claims[claimTokenVersion])
	if !ok {
		return nil, invalidClaim("invalid_tv", "tv claim parse failed")
	}

	jti, _ := claims["jti"].(string)

	issuedAt := iat.Time
	expiresAt := exp.Time

	return &Token{
		Token:        token,
		Issuer:       issuer,
		Audience:     []string(audience),
		Subject:      subject,
		ExtraClaims:  parseExtraClaims(claims),
		IssuedAt:     &issuedAt,
		ExpiresAt:    &expiresAt,
		ID:           jti,
		Type:         tokenType,
		TokenVersion: version,
	}, nil
}

func parseExtraClaims(claims jwtlib.MapClaims) ExtraClaims {
	var parsed ExtraClaims

	raw, ok := claims["extra_claims"].(map[string]interface{})
	if !ok {
		return parsed
	}

	rawRoles, _ := raw["roles"].([]interface{})
	for _, r := range rawRoles {
		if role, ok := r.(string); ok {
			parsed.Roles = append(parsed.Roles, role)
		}
	}

	return parsed
}

func parseTokenVersion(v any) (int, bool) {
	f, ok := v.(float64)
	if !ok || f < 0 || f != math.Trunc(f) || f > math.MaxInt32 {
		return 0, false
	}

	return int(f), true
}

func invalidClaim(code, msg string) error {
	return apperror.New(
		errors.New(code), msg, apperror.ParseKind,
		apperror.PackageOrigin, "jwt", nil, nil,
	)
}
