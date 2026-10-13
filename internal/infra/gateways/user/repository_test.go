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

func (s *UserRepositoryTestSuite) TestUpdatePassword() {
	s.Run("should update the user's password", func() {
		userID := uuid.New()

		s.mock.ExpectExec(regexp.QuoteMeta(queryUpdateUserPassword)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.repo.UpdatePassword(context.Background(), userID, "new-hash")

		s.Require().NoError(err)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when updating the password fails", func() {
		userID := uuid.New()

		s.mock.ExpectExec(regexp.QuoteMeta(queryUpdateUserPassword)).
			WillReturnError(sqlmock.ErrCancelled)

		err := s.repo.UpdatePassword(context.Background(), userID, "new-hash")

		s.Require().Error(err)
	})
}

func (s *UserRepositoryTestSuite) TestActivateUser() {
	s.Run("should activate the user", func() {
		userID := uuid.New()

		s.mock.ExpectExec(regexp.QuoteMeta(queryActivateUser)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.repo.ActivateUser(context.Background(), userID)

		s.Require().NoError(err)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when activating the user fails", func() {
		userID := uuid.New()

		s.mock.ExpectExec(regexp.QuoteMeta(queryActivateUser)).
			WillReturnError(sqlmock.ErrCancelled)

		err := s.repo.ActivateUser(context.Background(), userID)

		s.Require().Error(err)
	})
}

func (s *UserRepositoryTestSuite) TestSaveConfirmationCode() {
	s.Run("should save a confirmation code", func() {
		code := fixtures.InfraModelConfirmationCode(uuid.New())

		s.mock.ExpectExec(regexp.QuoteMeta(queryInsertConfirmationCode)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.repo.SaveConfirmationCode(context.Background(), code)

		s.Require().NoError(err)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when saving the confirmation code fails", func() {
		code := fixtures.InfraModelConfirmationCode(uuid.New())

		s.mock.ExpectExec(regexp.QuoteMeta(queryInsertConfirmationCode)).
			WillReturnError(sqlmock.ErrCancelled)

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

func (s *UserRepositoryTestSuite) TestMarkConfirmationCodeUsed() {
	s.Run("should mark the confirmation code as used", func() {
		codeID := uuid.New()

		s.mock.ExpectExec(regexp.QuoteMeta(queryMarkConfirmationCodeUsed)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.repo.MarkConfirmationCodeUsed(context.Background(), codeID)

		s.Require().NoError(err)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when marking the confirmation code as used fails", func() {
		codeID := uuid.New()

		s.mock.ExpectExec(regexp.QuoteMeta(queryMarkConfirmationCodeUsed)).
			WillReturnError(sqlmock.ErrCancelled)

		err := s.repo.MarkConfirmationCodeUsed(context.Background(), codeID)

		s.Require().Error(err)
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
	s.Run("should mark the refresh token as used", func() {
		id := uuid.New()

		s.mock.ExpectExec(regexp.QuoteMeta(queryMarkRefreshTokenUsed)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.repo.MarkRefreshTokenUsed(context.Background(), id)

		s.Require().NoError(err)
		s.NoError(s.mock.ExpectationsWereMet())
	})

	s.Run("should return an error when marking the refresh token as used fails", func() {
		id := uuid.New()

		s.mock.ExpectExec(regexp.QuoteMeta(queryMarkRefreshTokenUsed)).
			WillReturnError(sqlmock.ErrCancelled)

		err := s.repo.MarkRefreshTokenUsed(context.Background(), id)

		s.Require().Error(err)
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
