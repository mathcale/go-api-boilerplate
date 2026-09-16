package user

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/infra/database/models"
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
	s.mock.ExpectQuery(regexp.QuoteMeta(queryUserExistsIncludingInactive)).
		WithArgs("jane@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := s.repo.ExistsIncludingInactive(context.Background(), "jane@example.com")

	s.Require().NoError(err)
	s.Require().NotNil(exists)
	s.True(*exists)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *UserRepositoryTestSuite) TestGetByID_NotFound() {
	id := uuid.New()

	s.mock.ExpectQuery(regexp.QuoteMeta(queryGetUserByID)).
		WithArgs(id).
		WillReturnError(sqlmock.ErrCancelled)

	u, err := s.repo.GetByID(context.Background(), id)

	s.Require().Error(err)
	s.Nil(u)
}

func (s *UserRepositoryTestSuite) TestSave_CommitsUserAndCode() {
	u := models.User{
		ID:    uuid.New(),
		Email: "jane@example.com",
	}

	code := models.ConfirmationCode{
		ID:      uuid.New(),
		UserID:  u.ID,
		Purpose: "account_confirmation",
	}

	s.mock.ExpectBegin()

	s.mock.ExpectExec(regexp.QuoteMeta(queryInsertUser)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s.mock.ExpectExec(regexp.QuoteMeta(queryInsertConfirmationCode)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s.mock.ExpectCommit()

	err := s.repo.Save(context.Background(), u, code)

	s.Require().NoError(err)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *UserRepositoryTestSuite) TestSaveRefreshToken() {
	rt := models.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		FamilyID:  uuid.New(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	s.mock.ExpectExec(regexp.QuoteMeta(queryInsertRefreshToken)).
		WithArgs(rt.ID, rt.UserID, rt.FamilyID, rt.ExpiresAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.repo.SaveRefreshToken(context.Background(), rt)

	s.Require().NoError(err)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *UserRepositoryTestSuite) TestGetRefreshToken_Found() {
	id := uuid.New()
	userID := uuid.New()
	familyID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour)
	createdAt := time.Now()

	rows := sqlmock.NewRows(
		[]string{"id", "user_id", "family_id", "expires_at", "used_at", "revoked_at", "created_at"},
	).AddRow(id, userID, familyID, expiresAt, nil, nil, createdAt)

	s.mock.ExpectQuery(regexp.QuoteMeta(queryGetRefreshToken)).
		WithArgs(id).
		WillReturnRows(rows)

	rt, err := s.repo.GetRefreshToken(context.Background(), id)

	s.Require().NoError(err)
	s.Require().NotNil(rt)
	s.Equal(id, rt.ID)
	s.Equal(familyID, rt.FamilyID)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *UserRepositoryTestSuite) TestGetRefreshToken_NotFound() {
	id := uuid.New()

	s.mock.ExpectQuery(regexp.QuoteMeta(queryGetRefreshToken)).
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)

	rt, err := s.repo.GetRefreshToken(context.Background(), id)

	s.Require().NoError(err)
	s.Nil(rt)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *UserRepositoryTestSuite) TestMarkRefreshTokenUsed() {
	id := uuid.New()

	s.mock.ExpectExec(regexp.QuoteMeta(queryMarkRefreshTokenUsed)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.repo.MarkRefreshTokenUsed(context.Background(), id)

	s.Require().NoError(err)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *UserRepositoryTestSuite) TestRevokeRefreshTokenFamily() {
	familyID := uuid.New()

	s.mock.ExpectExec(regexp.QuoteMeta(queryRevokeRefreshTokenFamily)).
		WillReturnResult(sqlmock.NewResult(1, 2))

	err := s.repo.RevokeRefreshTokenFamily(context.Background(), familyID)

	s.Require().NoError(err)
	s.NoError(s.mock.ExpectationsWereMet())
}
