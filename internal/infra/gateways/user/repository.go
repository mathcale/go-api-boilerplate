package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mathcale/go-api-boilerplate/internal/infra/database/models"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
)

const (
	queryUserExistsIncludingInactive     = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	queryGetUserByEmailIncludingInactive = `SELECT * FROM users WHERE email = $1`
	queryGetUserByID                     = `SELECT * FROM users WHERE id = $1`
	queryInsertUser                      = `INSERT INTO users (id, name, surname, avatar_url, email, password, active, roles) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	queryUpdateUserPassword              = `UPDATE users SET password = $1, updated_at = $2 WHERE id = $3`
	queryActivateUser                    = `UPDATE users SET active = TRUE, updated_at = $1 WHERE id = $2`
	queryInsertConfirmationCode          = `INSERT INTO account_confirmation_codes (id, user_id, code, purpose, expires_at) VALUES ($1, $2, $3, $4, $5)`
	queryGetConfirmationCode             = `SELECT * FROM account_confirmation_codes WHERE user_id = $1 AND code = $2 AND purpose = $3 AND used IS FALSE`
	queryMarkConfirmationCodeUsed        = `UPDATE account_confirmation_codes SET used = TRUE WHERE id = $1`
	queryInsertRefreshToken              = `INSERT INTO refresh_tokens (id, user_id, family_id, expires_at) VALUES ($1, $2, $3, $4)`
	queryGetRefreshToken                 = `SELECT * FROM refresh_tokens WHERE id = $1`
	queryMarkRefreshTokenUsed            = `UPDATE refresh_tokens SET used_at = $1 WHERE id = $2`
	queryRevokeRefreshTokenFamily        = `UPDATE refresh_tokens SET revoked_at = $1 WHERE family_id = $2 AND revoked_at IS NULL`
)

type (
	Repository interface {
		ExistsIncludingInactive(ctx context.Context, email string) (*bool, error)
		GetIncludingInactive(ctx context.Context, email string) (*models.User, error)
		GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
		Save(ctx context.Context, user models.User, code models.ConfirmationCode) error
		UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
		ActivateUser(ctx context.Context, userID uuid.UUID) error
		SaveConfirmationCode(ctx context.Context, code models.ConfirmationCode) error
		GetConfirmationCode(
			ctx context.Context,
			userID uuid.UUID,
			code string,
			purpose string,
		) (*models.ConfirmationCode, error)
		MarkConfirmationCodeUsed(ctx context.Context, codeID uuid.UUID) error
		SaveRefreshToken(ctx context.Context, token models.RefreshToken) error
		GetRefreshToken(ctx context.Context, tokenID uuid.UUID) (*models.RefreshToken, error)
		MarkRefreshTokenUsed(ctx context.Context, tokenID uuid.UUID) error
		RevokeRefreshTokenFamily(ctx context.Context, familyID uuid.UUID) error
	}

	repository struct {
		db *sqlx.DB
	}
)

func NewUserRepository(db *sqlx.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) ExistsIncludingInactive(ctx context.Context, email string) (*bool, error) {
	var exists bool

	if err := r.db.GetContext(ctx, &exists, queryUserExistsIncludingInactive, email); err != nil {
		return nil, apperror.New(
			err, "user exists query failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	return &exists, nil
}

func (r *repository) GetIncludingInactive(ctx context.Context, email string) (*models.User, error) {
	var u models.User

	if err := r.db.GetContext(ctx, &u, queryGetUserByEmailIncludingInactive, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, apperror.New(
			err, "get user by email query failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	return &u, nil
}

func (r *repository) GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var u models.User

	if err := r.db.GetContext(ctx, &u, queryGetUserByID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, apperror.New(
			err, "get user by id query failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	return &u, nil
}

func (r *repository) Save(ctx context.Context, u models.User, code models.ConfirmationCode) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return apperror.New(
			err, "begin transaction failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	if _, err := tx.ExecContext(
		ctx,
		queryInsertUser,
		u.ID,
		u.Name,
		u.Surname,
		u.AvatarURL,
		u.Email,
		u.Password,
		u.Active,
		u.Roles,
	); err != nil {
		return r.rollback(tx, err, "insert user query failed")
	}

	if _, err := tx.ExecContext(
		ctx,
		queryInsertConfirmationCode,
		code.ID,
		code.UserID,
		code.Code,
		code.Purpose,
		code.ExpiresAt,
	); err != nil {
		return r.rollback(tx, err, "insert confirmation code query failed")
	}

	if err := tx.Commit(); err != nil {
		return apperror.New(
			err, "commit transaction failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	return nil
}

func (r *repository) UpdatePassword(
	ctx context.Context,
	userID uuid.UUID,
	hashedPassword string,
) error {
	if _, err := r.db.ExecContext(
		ctx,
		queryUpdateUserPassword,
		hashedPassword,
		time.Now(),
		userID,
	); err != nil {
		return apperror.New(
			err, "update user password query failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	return nil
}

func (r *repository) ActivateUser(ctx context.Context, userID uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, queryActivateUser, time.Now(), userID); err != nil {
		return apperror.New(
			err, "activate user query failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	return nil
}

func (r *repository) SaveConfirmationCode(ctx context.Context, code models.ConfirmationCode) error {
	if _, err := r.db.ExecContext(
		ctx,
		queryInsertConfirmationCode,
		code.ID,
		code.UserID,
		code.Code,
		code.Purpose,
		code.ExpiresAt,
	); err != nil {
		return apperror.New(
			err, "insert confirmation code query failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	return nil
}

func (r *repository) GetConfirmationCode(
	ctx context.Context,
	userID uuid.UUID,
	code string,
	purpose string,
) (*models.ConfirmationCode, error) {
	var cc models.ConfirmationCode

	if err := r.db.GetContext(
		ctx,
		&cc,
		queryGetConfirmationCode,
		userID,
		code,
		purpose,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, apperror.New(
			err, "get confirmation code query failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	return &cc, nil
}

func (r *repository) MarkConfirmationCodeUsed(ctx context.Context, codeID uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, queryMarkConfirmationCodeUsed, codeID); err != nil {
		return apperror.New(
			err, "mark confirmation code used query failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	return nil
}

func (r *repository) SaveRefreshToken(ctx context.Context, t models.RefreshToken) error {
	if _, err := r.db.ExecContext(
		ctx, queryInsertRefreshToken, t.ID, t.UserID, t.FamilyID, t.ExpiresAt,
	); err != nil {
		return apperror.New(
			err, "insert refresh token query failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	return nil
}

func (r *repository) GetRefreshToken(
	ctx context.Context,
	tokenID uuid.UUID,
) (*models.RefreshToken, error) {
	var t models.RefreshToken

	if err := r.db.GetContext(ctx, &t, queryGetRefreshToken, tokenID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, apperror.New(
			err, "get refresh token query failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	return &t, nil
}

func (r *repository) MarkRefreshTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, queryMarkRefreshTokenUsed, time.Now(), tokenID); err != nil {
		return apperror.New(
			err, "mark refresh token used query failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	return nil
}

func (r *repository) RevokeRefreshTokenFamily(ctx context.Context, familyID uuid.UUID) error {
	if _, err := r.db.ExecContext(
		ctx, queryRevokeRefreshTokenFamily, time.Now(), familyID,
	); err != nil {
		return apperror.New(
			err, "revoke refresh token family query failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	return nil
}

func (r *repository) rollback(tx *sqlx.Tx, cause error, msg string) error {
	if rbErr := tx.Rollback(); rbErr != nil {
		return apperror.New(
			errors.Join(cause, rbErr), msg+" (rollback also failed)", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	return apperror.New(
		cause, msg, apperror.DatabaseKind,
		apperror.RepositoryOrigin, "user", nil, nil,
	)
}
