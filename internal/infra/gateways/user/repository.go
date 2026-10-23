package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"go-api-boilerplate/internal/infra/database/models"
	"go-api-boilerplate/internal/pkg/apperror"
)

const (
	queryUserExistsIncludingInactive     = `SELECT EXISTS(SELECT 1 FROM users WHERE lower(email) = lower($1))`
	queryGetUserByEmailIncludingInactive = `SELECT * FROM users WHERE lower(email) = lower($1)`
	queryGetUserByID                     = `SELECT * FROM users WHERE id = $1`
	queryInsertUser                      = `INSERT INTO users (id, name, surname, avatar_url, email, password, active, roles) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	queryGetUserAuthState                = `SELECT token_version, active FROM users WHERE id = $1`
	queryUpdateUserPassword              = `UPDATE users SET password = $1, token_version = token_version + 1, updated_at = $2 WHERE id = $3` // #nosec G101 -- SQL statement, not a credential
	queryActivateUser                    = `UPDATE users SET active = TRUE, updated_at = $1 WHERE id = $2`
	queryInsertConfirmationCode          = `INSERT INTO account_confirmation_codes (id, user_id, code, purpose, expires_at) VALUES ($1, $2, $3, $4, $5)`
	queryGetConfirmationCode             = `SELECT * FROM account_confirmation_codes WHERE user_id = $1 AND code = $2 AND purpose = $3 AND used IS FALSE`
	queryConsumeConfirmationCode         = `UPDATE account_confirmation_codes SET used = TRUE WHERE id = $1 AND used IS FALSE AND expires_at > $2`
	queryInvalidateConfirmationCodes     = `UPDATE account_confirmation_codes SET used = TRUE WHERE user_id = $1 AND purpose = $2 AND used IS FALSE`
	queryInsertRefreshToken              = `INSERT INTO refresh_tokens (id, user_id, family_id, expires_at) VALUES ($1, $2, $3, $4)` // #nosec G101 -- SQL statement, not a credential
	queryGetRefreshToken                 = `SELECT * FROM refresh_tokens WHERE id = $1`                                              // #nosec G101 -- SQL statement, not a credential
	queryMarkRefreshTokenUsed            = `UPDATE refresh_tokens SET used_at = $1 WHERE id = $2 AND used_at IS NULL AND revoked_at IS NULL`
	queryRevokeRefreshTokenFamily        = `UPDATE refresh_tokens SET revoked_at = $1 WHERE family_id = $2 AND revoked_at IS NULL`
	queryRevokeUserRefreshTokens         = `UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`
)

type (
	Repository interface {
		ExistsIncludingInactive(ctx context.Context, email string) (*bool, error)
		GetIncludingInactive(ctx context.Context, email string) (*models.User, error)
		GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
		GetAuthState(ctx context.Context, userID uuid.UUID) (*models.UserAuthState, error)
		Save(ctx context.Context, user models.User, code models.ConfirmationCode) error
		UpdatePasswordWithCode(
			ctx context.Context,
			userID uuid.UUID,
			codeID uuid.UUID,
			hashedPassword string,
		) (bool, error)
		ActivateUserWithCode(ctx context.Context, userID uuid.UUID, codeID uuid.UUID) (bool, error)
		SaveConfirmationCode(ctx context.Context, code models.ConfirmationCode) error
		GetConfirmationCode(
			ctx context.Context,
			userID uuid.UUID,
			code string,
			purpose string,
		) (*models.ConfirmationCode, error)
		SaveRefreshToken(ctx context.Context, token models.RefreshToken) error
		GetRefreshToken(ctx context.Context, tokenID uuid.UUID) (*models.RefreshToken, error)
		MarkRefreshTokenUsed(ctx context.Context, tokenID uuid.UUID) (bool, error)
		RevokeRefreshTokenFamily(ctx context.Context, familyID uuid.UUID) error
		RevokeUserRefreshTokens(ctx context.Context, userID uuid.UUID) error
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

func (r *repository) GetAuthState(
	ctx context.Context,
	userID uuid.UUID,
) (*models.UserAuthState, error) {
	var state models.UserAuthState

	if err := r.db.GetContext(ctx, &state, queryGetUserAuthState, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, apperror.New(
			err, "get user auth state query failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	return &state, nil
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

func (r *repository) UpdatePasswordWithCode(
	ctx context.Context,
	userID uuid.UUID,
	codeID uuid.UUID,
	hashedPassword string,
) (bool, error) {
	now := time.Now()

	return r.consumeCodeAndExec(
		ctx, codeID, now, "update user password",
		queryUpdateUserPassword, hashedPassword, now, userID,
	)
}

func (r *repository) ActivateUserWithCode(
	ctx context.Context,
	userID uuid.UUID,
	codeID uuid.UUID,
) (bool, error) {
	now := time.Now()

	return r.consumeCodeAndExec(
		ctx, codeID, now, "activate user", queryActivateUser, now, userID,
	)
}

func (r *repository) SaveConfirmationCode(ctx context.Context, code models.ConfirmationCode) error {
	tx, err := r.begin(ctx)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx, queryInvalidateConfirmationCodes, code.UserID, code.Purpose,
	); err != nil {
		return r.rollback(tx, err, "invalidate confirmation codes query failed")
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

	return r.commit(tx)
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

func (r *repository) MarkRefreshTokenUsed(ctx context.Context, tokenID uuid.UUID) (bool, error) {
	res, err := r.db.ExecContext(ctx, queryMarkRefreshTokenUsed, time.Now(), tokenID)
	if err != nil {
		return false, apperror.New(
			err, "mark refresh token used query failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, apperror.New(
			err, "mark refresh token used rows affected failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	return affected > 0, nil
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

func (r *repository) RevokeUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	if _, err := r.db.ExecContext(
		ctx, queryRevokeUserRefreshTokens, time.Now(), userID,
	); err != nil {
		return apperror.New(
			err, "revoke user refresh tokens query failed", apperror.DatabaseKind,
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

func (r *repository) consumeCodeAndExec(
	ctx context.Context,
	codeID uuid.UUID,
	now time.Time,
	action string,
	query string,
	args ...any,
) (bool, error) {
	tx, err := r.begin(ctx)
	if err != nil {
		return false, err
	}

	res, err := tx.ExecContext(ctx, queryConsumeConfirmationCode, codeID, now)
	if err != nil {
		return false, r.rollback(tx, err, "consume confirmation code query failed")
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, r.rollback(tx, err, "consume confirmation code rows affected failed")
	}

	if affected == 0 {
		if rbErr := tx.Rollback(); rbErr != nil {
			return false, apperror.New(
				rbErr, "rollback after lost code race failed", apperror.DatabaseKind,
				apperror.RepositoryOrigin, "user", nil, nil,
			)
		}

		return false, nil
	}

	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return false, r.rollback(tx, err, action+" query failed")
	}

	if err := r.commit(tx); err != nil {
		return false, err
	}

	return true, nil
}

func (r *repository) begin(ctx context.Context) (*sqlx.Tx, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, apperror.New(
			err, "begin transaction failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	return tx, nil
}

func (r *repository) commit(tx *sqlx.Tx) error {
	if err := tx.Commit(); err != nil {
		return apperror.New(
			err, "commit transaction failed", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "user", nil, nil,
		)
	}

	return nil
}
