package user

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/tests/fixtures"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo Repository
}

func (s *UserRepositoryTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	s.Require().NoError(err)

	s.db = db
	s.mock = mock
	s.repo = NewUserRepository(sqlx.NewDb(db, "sqlmock"))
}

func (s *UserRepositoryTestSuite) TearDownTest() {
	_ = s.db.Close()
}

func TestUserRepository(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (s *UserRepositoryTestSuite) TestExistsIncludingInactive() {
	s.Run("should return true when a user exists including inactive users", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(queryUserExistsIncludingInactive)).
			WithArgs(fixtures.InfraUserEmail).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		exists, err := s.repo.ExistsIncludingInactive(context.Background(), fixtures.InfraUserEmail)

		s.Require().NoError(err)
		s.Require().NotNil(exists)
		s.True(*exists)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when the query fails", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(queryUserExistsIncludingInactive)).
			WithArgs(fixtures.InfraUserEmail).
			WillReturnError(sqlmock.ErrCancelled)

		exists, err := s.repo.ExistsIncludingInactive(context.Background(), fixtures.InfraUserEmail)

		s.Require().Error(err)
		s.Nil(exists)
	})
}

func (s *UserRepositoryTestSuite) TestGetByID() {
	s.Run("should return the user when found by id", func() {
		id := uuid.New()

		s.mock.ExpectQuery(regexp.QuoteMeta(queryGetUserByID)).
			WithArgs(id).
			WillReturnRows(fixtures.InfraUserRows(id))

		u, err := s.repo.GetByID(context.Background(), id)

		s.Require().NoError(err)
		s.Require().NotNil(u)
		s.Equal(id, u.ID)
		s.Equal(4, u.TokenVersion)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return nil when no user is found by id", func() {
		id := uuid.New()

		s.mock.ExpectQuery(regexp.QuoteMeta(queryGetUserByID)).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		u, err := s.repo.GetByID(context.Background(), id)

		s.Require().NoError(err)
		s.Nil(u)
	})

	s.Run("should return an error when the query fails", func() {
		id := uuid.New()

		s.mock.ExpectQuery(regexp.QuoteMeta(queryGetUserByID)).
			WithArgs(id).
			WillReturnError(sqlmock.ErrCancelled)

		u, err := s.repo.GetByID(context.Background(), id)

		s.Require().Error(err)
		s.Nil(u)
	})
}

func (s *UserRepositoryTestSuite) TestGetAuthState() {
	s.Run("should return the token version and active flag of the user", func() {
		id := uuid.New()

		s.mock.ExpectQuery(regexp.QuoteMeta(queryGetUserAuthState)).
			WithArgs(id).
			WillReturnRows(fixtures.InfraUserAuthStateRows(5, true))

		state, err := s.repo.GetAuthState(context.Background(), id)

		s.Require().NoError(err)
		s.Require().NotNil(state)
		s.Equal(5, state.TokenVersion)
		s.True(state.Active)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return nil when the user does not exist", func() {
		id := uuid.New()

		s.mock.ExpectQuery(regexp.QuoteMeta(queryGetUserAuthState)).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		state, err := s.repo.GetAuthState(context.Background(), id)

		s.Require().NoError(err)
		s.Nil(state)
	})

	s.Run("should return an error when the query fails", func() {
		id := uuid.New()

		s.mock.ExpectQuery(regexp.QuoteMeta(queryGetUserAuthState)).
			WithArgs(id).
			WillReturnError(sqlmock.ErrCancelled)

		state, err := s.repo.GetAuthState(context.Background(), id)

		s.Require().Error(err)
		s.Nil(state)
	})
}

func (s *UserRepositoryTestSuite) TestGetIncludingInactive() {
	s.Run("should return the user when found by email including inactive users", func() {
		id := uuid.New()

		s.mock.ExpectQuery(regexp.QuoteMeta(queryGetUserByEmailIncludingInactive)).
			WithArgs(fixtures.InfraUserEmail).
			WillReturnRows(fixtures.InfraUserRows(id))

		u, err := s.repo.GetIncludingInactive(context.Background(), fixtures.InfraUserEmail)

		s.Require().NoError(err)
		s.Require().NotNil(u)
		s.Equal(id, u.ID)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return nil when no user is found by email", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(queryGetUserByEmailIncludingInactive)).
			WithArgs(fixtures.InfraUnknownEmail).
			WillReturnError(sql.ErrNoRows)

		u, err := s.repo.GetIncludingInactive(context.Background(), fixtures.InfraUnknownEmail)

		s.Require().NoError(err)
		s.Nil(u)
	})

	s.Run("should return an error when the query fails", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(queryGetUserByEmailIncludingInactive)).
			WithArgs(fixtures.InfraUserEmail).
			WillReturnError(sqlmock.ErrCancelled)

		u, err := s.repo.GetIncludingInactive(context.Background(), fixtures.InfraUserEmail)

		s.Require().Error(err)
		s.Nil(u)
	})
}

func (s *UserRepositoryTestSuite) TestSave() {
	s.Run("should insert the user and confirmation code and commit the transaction", func() {
		u := fixtures.InfraModelUser()
		code := fixtures.InfraModelConfirmationCode(u.ID)

		s.mock.ExpectBegin()

		s.mock.ExpectExec(regexp.QuoteMeta(queryInsertUser)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		s.mock.ExpectExec(regexp.QuoteMeta(queryInsertConfirmationCode)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		s.mock.ExpectCommit()

		err := s.repo.Save(context.Background(), u, code)

		s.Require().NoError(err)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when beginning the transaction fails", func() {
		u := fixtures.InfraModelUser()
		code := fixtures.InfraModelConfirmationCode(u.ID)

		s.mock.ExpectBegin().WillReturnError(sqlmock.ErrCancelled)

		err := s.repo.Save(context.Background(), u, code)

		s.Require().Error(err)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should roll back the transaction when inserting the user fails", func() {
		u := fixtures.InfraModelUser()
		code := fixtures.InfraModelConfirmationCode(u.ID)

		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(queryInsertUser)).WillReturnError(sqlmock.ErrCancelled)
		s.mock.ExpectRollback()

		err := s.repo.Save(context.Background(), u, code)

		s.Require().Error(err)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when both the user insert and the rollback fail", func() {
		u := fixtures.InfraModelUser()
		code := fixtures.InfraModelConfirmationCode(u.ID)

		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(queryInsertUser)).WillReturnError(sqlmock.ErrCancelled)
		s.mock.ExpectRollback().WillReturnError(errors.New("rollback failed"))

		err := s.repo.Save(context.Background(), u, code)

		s.Require().Error(err)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should roll back the transaction when inserting the confirmation code fails", func() {
		u := fixtures.InfraModelUser()
		code := fixtures.InfraModelConfirmationCode(u.ID)

		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(queryInsertUser)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		s.mock.ExpectExec(regexp.QuoteMeta(queryInsertConfirmationCode)).
			WillReturnError(sqlmock.ErrCancelled)
		s.mock.ExpectRollback()

		err := s.repo.Save(context.Background(), u, code)

		s.Require().Error(err)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when committing the transaction fails", func() {
		u := fixtures.InfraModelUser()
		code := fixtures.InfraModelConfirmationCode(u.ID)

		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(queryInsertUser)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		s.mock.ExpectExec(regexp.QuoteMeta(queryInsertConfirmationCode)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		s.mock.ExpectCommit().WillReturnError(sqlmock.ErrCancelled)

		err := s.repo.Save(context.Background(), u, code)

		s.Require().Error(err)
		s.NoError(s.mock.ExpectationsWereMet())
	})
}

func (s *UserRepositoryTestSuite) TestUpdatePasswordWithCode() {
	s.Run("should consume the code and update the password in one transaction", func() {
		userID := uuid.New()
		codeID := uuid.New()

		s.Contains(queryUpdateUserPassword, "token_version = token_version + 1")

		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(queryConsumeConfirmationCode)).
			WithArgs(codeID, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
		s.mock.ExpectExec(regexp.QuoteMeta(queryUpdateUserPassword)).
			WithArgs("new-hash", sqlmock.AnyArg(), userID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		s.mock.ExpectCommit()

		updated, err := s.repo.UpdatePasswordWithCode(
			context.Background(), userID, codeID, "new-hash",
		)

		s.Require().NoError(err)
		s.True(updated)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should roll back without updating when the code was already consumed", func() {
		userID := uuid.New()
		codeID := uuid.New()

		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(queryConsumeConfirmationCode)).
			WithArgs(codeID, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0))
		s.mock.ExpectRollback()

		updated, err := s.repo.UpdatePasswordWithCode(
			context.Background(), userID, codeID, "new-hash",
		)

		s.Require().NoError(err)
		s.False(updated)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when beginning the transaction fails", func() {
		s.mock.ExpectBegin().WillReturnError(sqlmock.ErrCancelled)

		updated, err := s.repo.UpdatePasswordWithCode(
			context.Background(), uuid.New(), uuid.New(), "new-hash",
		)

		s.Require().Error(err)
		s.False(updated)
	})

	s.Run("should roll back when consuming the code fails", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(queryConsumeConfirmationCode)).
			WillReturnError(sqlmock.ErrCancelled)
		s.mock.ExpectRollback()

		updated, err := s.repo.UpdatePasswordWithCode(
			context.Background(), uuid.New(), uuid.New(), "new-hash",
		)

		s.Require().Error(err)
		s.False(updated)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should roll back when updating the password fails", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(queryConsumeConfirmationCode)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		s.mock.ExpectExec(regexp.QuoteMeta(queryUpdateUserPassword)).
			WillReturnError(sqlmock.ErrCancelled)
		s.mock.ExpectRollback()

		updated, err := s.repo.UpdatePasswordWithCode(
			context.Background(), uuid.New(), uuid.New(), "new-hash",
		)

		s.Require().Error(err)
		s.False(updated)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when the commit fails", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(queryConsumeConfirmationCode)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		s.mock.ExpectExec(regexp.QuoteMeta(queryUpdateUserPassword)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		s.mock.ExpectCommit().WillReturnError(sqlmock.ErrCancelled)

		updated, err := s.repo.UpdatePasswordWithCode(
			context.Background(), uuid.New(), uuid.New(), "new-hash",
		)

		s.Require().Error(err)
		s.False(updated)
	})
}

func (s *UserRepositoryTestSuite) TestActivateUserWithCode() {
	s.Run("should consume the code and activate the user in one transaction", func() {
		userID := uuid.New()
		codeID := uuid.New()

		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(queryConsumeConfirmationCode)).
			WithArgs(codeID, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
		s.mock.ExpectExec(regexp.QuoteMeta(queryActivateUser)).
			WithArgs(sqlmock.AnyArg(), userID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		s.mock.ExpectCommit()

		activated, err := s.repo.ActivateUserWithCode(context.Background(), userID, codeID)

		s.Require().NoError(err)
		s.True(activated)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should roll back without activating when the code was already consumed", func() {
		userID := uuid.New()
		codeID := uuid.New()

		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(queryConsumeConfirmationCode)).
			WithArgs(codeID, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0))
		s.mock.ExpectRollback()

		activated, err := s.repo.ActivateUserWithCode(context.Background(), userID, codeID)

		s.Require().NoError(err)
		s.False(activated)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when beginning the transaction fails", func() {
		s.mock.ExpectBegin().WillReturnError(sqlmock.ErrCancelled)

		activated, err := s.repo.ActivateUserWithCode(
			context.Background(), uuid.New(), uuid.New(),
		)

		s.Require().Error(err)
		s.False(activated)
	})

	s.Run("should roll back when activating the user fails", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(queryConsumeConfirmationCode)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		s.mock.ExpectExec(regexp.QuoteMeta(queryActivateUser)).
			WillReturnError(sqlmock.ErrCancelled)
		s.mock.ExpectRollback()

		activated, err := s.repo.ActivateUserWithCode(
			context.Background(), uuid.New(), uuid.New(),
		)

		s.Require().Error(err)
		s.False(activated)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should roll back when consuming the code fails", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(queryConsumeConfirmationCode)).
			WillReturnError(sqlmock.ErrCancelled)
		s.mock.ExpectRollback()

		activated, err := s.repo.ActivateUserWithCode(
			context.Background(), uuid.New(), uuid.New(),
		)

		s.Require().Error(err)
		s.False(activated)
		s.NoError(s.mock.ExpectationsWereMet())
	})
}

func (s *UserRepositoryTestSuite) TestSaveConfirmationCode() {
	s.Run("should invalidate previous codes and insert the new one in a transaction", func() {
		code := fixtures.InfraModelConfirmationCode(uuid.New())

		s.Contains(queryInvalidateConfirmationCodes, "used IS FALSE")

		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(queryInvalidateConfirmationCodes)).
			WithArgs(code.UserID, code.Purpose).
			WillReturnResult(sqlmock.NewResult(0, 2))
		s.mock.ExpectExec(regexp.QuoteMeta(queryInsertConfirmationCode)).
			WithArgs(code.ID, code.UserID, code.Code, code.Purpose, code.ExpiresAt).
			WillReturnResult(sqlmock.NewResult(1, 1))
		s.mock.ExpectCommit()

		err := s.repo.SaveConfirmationCode(context.Background(), code)

		s.Require().NoError(err)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when beginning the transaction fails", func() {
		code := fixtures.InfraModelConfirmationCode(uuid.New())

		s.mock.ExpectBegin().WillReturnError(sqlmock.ErrCancelled)

		err := s.repo.SaveConfirmationCode(context.Background(), code)

		s.Require().Error(err)
	})

	s.Run("should roll back when invalidating previous codes fails", func() {
		code := fixtures.InfraModelConfirmationCode(uuid.New())

		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(queryInvalidateConfirmationCodes)).
			WillReturnError(sqlmock.ErrCancelled)
		s.mock.ExpectRollback()

		err := s.repo.SaveConfirmationCode(context.Background(), code)

		s.Require().Error(err)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should roll back when inserting the confirmation code fails", func() {
		code := fixtures.InfraModelConfirmationCode(uuid.New())

		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(queryInvalidateConfirmationCodes)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		s.mock.ExpectExec(regexp.QuoteMeta(queryInsertConfirmationCode)).
			WillReturnError(sqlmock.ErrCancelled)
		s.mock.ExpectRollback()

		err := s.repo.SaveConfirmationCode(context.Background(), code)

		s.Require().Error(err)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when the commit fails", func() {
		code := fixtures.InfraModelConfirmationCode(uuid.New())

		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(queryInvalidateConfirmationCodes)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		s.mock.ExpectExec(regexp.QuoteMeta(queryInsertConfirmationCode)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		s.mock.ExpectCommit().WillReturnError(sqlmock.ErrCancelled)

		err := s.repo.SaveConfirmationCode(context.Background(), code)

		s.Require().Error(err)
	})
}

func (s *UserRepositoryTestSuite) TestGetConfirmationCode() {
	s.Run("should return the confirmation code when found", func() {
		userID := uuid.New()
		id := uuid.New()

		s.mock.ExpectQuery(regexp.QuoteMeta(queryGetConfirmationCode)).
			WithArgs(userID, "the-code", "account_confirmation").
			WillReturnRows(fixtures.InfraConfirmationCodeRows(id, userID, "the-code"))

		cc, err := s.repo.GetConfirmationCode(
			context.Background(),
			userID,
			"the-code",
			"account_confirmation",
		)

		s.Require().NoError(err)
		s.Require().NotNil(cc)
		s.Equal(id, cc.ID)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return nil when no confirmation code is found", func() {
		userID := uuid.New()

		s.mock.ExpectQuery(regexp.QuoteMeta(queryGetConfirmationCode)).
			WithArgs(userID, "nope", "account_confirmation").
			WillReturnError(sql.ErrNoRows)

		cc, err := s.repo.GetConfirmationCode(
			context.Background(),
			userID,
			"nope",
			"account_confirmation",
		)

		s.Require().NoError(err)
		s.Nil(cc)
	})

	s.Run("should return an error when the query fails", func() {
		userID := uuid.New()

		s.mock.ExpectQuery(regexp.QuoteMeta(queryGetConfirmationCode)).
			WithArgs(userID, "the-code", "account_confirmation").
			WillReturnError(sqlmock.ErrCancelled)

		cc, err := s.repo.GetConfirmationCode(
			context.Background(),
			userID,
			"the-code",
			"account_confirmation",
		)

		s.Require().Error(err)
		s.Nil(cc)
	})
}

func (s *UserRepositoryTestSuite) TestSaveRefreshToken() {
	s.Run("should save a new refresh token", func() {
		rt := fixtures.InfraModelRefreshToken()

		s.mock.ExpectExec(regexp.QuoteMeta(queryInsertRefreshToken)).
			WithArgs(rt.ID, rt.UserID, rt.FamilyID, rt.ExpiresAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.repo.SaveRefreshToken(context.Background(), rt)

		s.Require().NoError(err)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when saving the refresh token fails", func() {
		rt := fixtures.InfraModelRefreshToken()

		s.mock.ExpectExec(regexp.QuoteMeta(queryInsertRefreshToken)).
			WillReturnError(sqlmock.ErrCancelled)

		err := s.repo.SaveRefreshToken(context.Background(), rt)

		s.Require().Error(err)
	})
}

func (s *UserRepositoryTestSuite) TestGetRefreshToken() {
	s.Run("should return the refresh token when found", func() {
		id := uuid.New()
		familyID := uuid.New()

		s.mock.ExpectQuery(regexp.QuoteMeta(queryGetRefreshToken)).
			WithArgs(id).
			WillReturnRows(fixtures.InfraRefreshTokenRows(id, uuid.New(), familyID))

		rt, err := s.repo.GetRefreshToken(context.Background(), id)

		s.Require().NoError(err)
		s.Require().NotNil(rt)
		s.Equal(id, rt.ID)
		s.Equal(familyID, rt.FamilyID)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return nil when no refresh token is found", func() {
		id := uuid.New()

		s.mock.ExpectQuery(regexp.QuoteMeta(queryGetRefreshToken)).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		rt, err := s.repo.GetRefreshToken(context.Background(), id)

		s.Require().NoError(err)
		s.Nil(rt)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when the query fails", func() {
		id := uuid.New()

		s.mock.ExpectQuery(regexp.QuoteMeta(queryGetRefreshToken)).
			WithArgs(id).
			WillReturnError(sqlmock.ErrCancelled)

		rt, err := s.repo.GetRefreshToken(context.Background(), id)

		s.Require().Error(err)
		s.Nil(rt)
	})
}

func (s *UserRepositoryTestSuite) TestMarkRefreshTokenUsed() {
	s.Run("should report the token as claimed when a row is affected", func() {
		id := uuid.New()

		s.Contains(queryMarkRefreshTokenUsed, "used_at IS NULL AND revoked_at IS NULL")

		s.mock.ExpectExec(regexp.QuoteMeta(queryMarkRefreshTokenUsed)).
			WithArgs(sqlmock.AnyArg(), id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		claimed, err := s.repo.MarkRefreshTokenUsed(context.Background(), id)

		s.Require().NoError(err)
		s.True(claimed)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should report the token as not claimed when no row is affected", func() {
		id := uuid.New()

		s.mock.ExpectExec(regexp.QuoteMeta(queryMarkRefreshTokenUsed)).
			WithArgs(sqlmock.AnyArg(), id).
			WillReturnResult(sqlmock.NewResult(0, 0))

		claimed, err := s.repo.MarkRefreshTokenUsed(context.Background(), id)

		s.Require().NoError(err)
		s.False(claimed)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when marking the refresh token as used fails", func() {
		id := uuid.New()

		s.mock.ExpectExec(regexp.QuoteMeta(queryMarkRefreshTokenUsed)).
			WillReturnError(sqlmock.ErrCancelled)

		claimed, err := s.repo.MarkRefreshTokenUsed(context.Background(), id)

		s.Require().Error(err)
		s.False(claimed)
	})

	s.Run("should return an error when rows affected cannot be read", func() {
		id := uuid.New()

		s.mock.ExpectExec(regexp.QuoteMeta(queryMarkRefreshTokenUsed)).
			WillReturnResult(sqlmock.NewErrorResult(sqlmock.ErrCancelled))

		claimed, err := s.repo.MarkRefreshTokenUsed(context.Background(), id)

		s.Require().Error(err)
		s.False(claimed)
	})
}

func (s *UserRepositoryTestSuite) TestRevokeRefreshTokenFamily() {
	s.Run("should revoke the entire refresh token family", func() {
		familyID := uuid.New()

		s.mock.ExpectExec(regexp.QuoteMeta(queryRevokeRefreshTokenFamily)).
			WillReturnResult(sqlmock.NewResult(1, 2))

		err := s.repo.RevokeRefreshTokenFamily(context.Background(), familyID)

		s.Require().NoError(err)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when revoking the refresh token family fails", func() {
		familyID := uuid.New()

		s.mock.ExpectExec(regexp.QuoteMeta(queryRevokeRefreshTokenFamily)).
			WillReturnError(sqlmock.ErrCancelled)

		err := s.repo.RevokeRefreshTokenFamily(context.Background(), familyID)

		s.Require().Error(err)
	})
}

func (s *UserRepositoryTestSuite) TestRevokeUserRefreshTokens() {
	s.Run("should revoke every refresh token of the user", func() {
		userID := uuid.New()

		s.mock.ExpectExec(regexp.QuoteMeta(queryRevokeUserRefreshTokens)).
			WithArgs(sqlmock.AnyArg(), userID).
			WillReturnResult(sqlmock.NewResult(1, 3))

		err := s.repo.RevokeUserRefreshTokens(context.Background(), userID)

		s.Require().NoError(err)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when revoking the user refresh tokens fails", func() {
		userID := uuid.New()

		s.mock.ExpectExec(regexp.QuoteMeta(queryRevokeUserRefreshTokens)).
			WillReturnError(sqlmock.ErrCancelled)

		err := s.repo.RevokeUserRefreshTokens(context.Background(), userID)

		s.Require().Error(err)
	})
}
