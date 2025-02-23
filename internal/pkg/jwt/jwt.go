package jwt

import (
	"errors"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type JWTAuth interface {
	IssueAccessToken(id string) (*Token, error)
	IssueRefreshToken(id string) (*Token, error)
	VerifyAccessToken(at string) (*Token, error)
	VerifyRefreshToken(rt string) (*Token, error)
}

type jwtAuth struct {
	logger          logger.Logger
	accessSecret    []byte
	accessLifetime  int
	refreshSecret   []byte
	refreshLifetime int
	issuer          string
	audience        string
	signingMethod   jwtlib.SigningMethod
}

type Token struct {
	Token     *string    `json:"token,omitempty"`
	Issuer    string     `json:"iss,omitempty"`
	Subject   string     `json:"sub,omitempty"`
	Audience  []string   `json:"aud,omitempty"`
	ExpiresAt *time.Time `json:"exp,omitempty"`
	IssuedAt  *time.Time `json:"iat,omitempty"`
}

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

func (j *jwtAuth) IssueAccessToken(id string) (*Token, error) {
	return j.issue(id, j.accessSecret, j.accessLifetime)
}

func (j *jwtAuth) VerifyAccessToken(at string) (*Token, error) {
	return j.verify(at, j.accessSecret)
}

func (j *jwtAuth) IssueRefreshToken(id string) (*Token, error) {
	return j.issue(id, j.refreshSecret, j.refreshLifetime)
}

func (j *jwtAuth) VerifyRefreshToken(rt string) (*Token, error) {
	return j.verify(rt, j.refreshSecret)
}

func (j *jwtAuth) issue(id string, secret []byte, lifetime int) (*Token, error) {
	claims := jwtlib.MapClaims{
		"sub": id,
		"iss": j.issuer,
		"aud": j.audience,
		"exp": time.Now().Add(time.Duration(lifetime) * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwtlib.NewWithClaims(j.signingMethod, claims)

	j.logger.Debug("Token claims added", map[string]interface{}{
		"claims": token,
	})

	tokenStr, err := token.SignedString(secret)
	if err != nil {
		return nil, err
	}

	return j.toToken(&tokenStr, claims)
}

func (j *jwtAuth) verify(token string, secret []byte) (*Token, error) {
	t, err := jwtlib.Parse(token, func(token *jwtlib.Token) (interface{}, error) {
		return secret, nil
	}, jwtlib.WithValidMethods([]string{j.signingMethod.Alg()}))
	if err != nil {
		return nil, err
	}

	claims, ok := t.Claims.(jwtlib.MapClaims)
	if !ok || !t.Valid {
		return nil, errors.New("invalid_token")
	}

	return j.toToken(&token, claims)
}

func (j *jwtAuth) toToken(token *string, claims jwtlib.MapClaims) (*Token, error) {
	iat, err := j.parseTime(claims["iat"])
	if err != nil {
		return nil, err
	}

	exp, err := j.parseTime(claims["exp"])
	if err != nil {
		return nil, err
	}

	return &Token{
		Token:     token,
		Issuer:    claims["iss"].(string),
		Audience:  []string{claims["aud"].(string)},
		Subject:   claims["sub"].(string),
		IssuedAt:  iat,
		ExpiresAt: exp,
	}, nil
}

func (j *jwtAuth) parseTime(v any) (*time.Time, error) {
	var t time.Time

	switch v.(type) {
	case int64:
		t = time.Unix(v.(int64), 0)
	case float64:
		t = time.Unix(int64(v.(float64)), 0)
	default:
		return nil, errors.New("invalid_time")
	}

	return &t, nil
}
