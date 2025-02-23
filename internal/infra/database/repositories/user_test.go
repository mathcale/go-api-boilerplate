package repositories

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/tests/fixtures"
)

type UserRepositoryTestSuite struct {
	suite.Suite

	conn *sql.DB
	db   *sqlx.DB
	mock sqlmock.Sqlmock
	ctx  context.Context

	repo User
}

func (s *UserRepositoryTestSuite) SetupSuite() {
	var err error

	s.conn, s.mock, err = sqlmock.New()
	if err != nil {
		s.FailNow(err.Error())
	}

	s.ctx = context.Background()
	s.db = sqlx.NewDb(s.conn, "sqlmock")
	s.repo = NewUserRepository(s.db)
}

func TestUserRepository(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (s *UserRepositoryTestSuite) TestExists() {
	domain := fixtures.NewUserDomain()

	s.Run("should return true when user exists", func() {
		rows := sqlmock.NewRows([]string{"exists"}).
			AddRow(true)

		s.mock.ExpectQuery(regexp.QuoteMeta(queryUserExists)).
			WithArgs(domain.Email).
			WillReturnRows(rows)

		exists, err := s.repo.Exists(s.ctx, domain.Email)

		s.NoError(err)
		s.True(*exists)
	})

	s.Run("should return error when something goes wrong while querying data", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(queryUserExists)).
			WithArgs(domain.Email).
			WillReturnError(sql.ErrConnDone)

		exists, err := s.repo.Exists(s.ctx, domain.Email)

		s.Error(err)
		s.Nil(exists)
	})
}

func (s *UserRepositoryTestSuite) TestGet() {
	fixt := fixtures.NewUserModel()

	s.Run("should return user by its email", func() {
		rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "active", "created_at", "updated_at"}).
			AddRow(fixt.ID, fixt.Name, fixt.Email, fixt.Password, fixt.Active, fixt.CreatedAt, fixt.UpdatedAt)

		s.mock.ExpectQuery(regexp.QuoteMeta(queryGetUserByEmail)).
			WithArgs(fixt.Email).
			WillReturnRows(rows)

		user, err := s.repo.Get(s.ctx, fixt.Email)

		s.NoError(err)
		s.Equal(fixt.ID, user.ID)
		s.Equal(fixt.Name, user.Name)
		s.Equal(fixt.Email, user.Email)
		s.Equal(fixt.Password, user.Password)
		s.Equal(fixt.Active, user.Active)
		s.Equal(fixt.CreatedAt, user.CreatedAt)
		s.Equal(fixt.UpdatedAt, user.UpdatedAt)
	})

	s.Run("should return error when something goes wrong while querying data", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(queryGetUserByEmail)).
			WithArgs(fixt.Email).
			WillReturnError(sql.ErrConnDone)

		user, err := s.repo.Get(s.ctx, fixt.Email)

		s.Error(err)
		s.Nil(user)
	})
}

func (s *UserRepositoryTestSuite) TestSave() {
	fixt := fixtures.NewUserModelComplete()

	s.Run("should save new user", func() {
		s.mock.ExpectExec(regexp.QuoteMeta(queryInsertUser)).
			WithArgs(fixt.Name, fixt.Email, fixt.Password, fixt.Active).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.repo.Save(s.ctx, fixt)

		s.NoError(err)
	})

	s.Run("should return error while saving user", func() {
		s.mock.ExpectExec(regexp.QuoteMeta(queryInsertUser)).
			WithArgs(fixt.Name, fixt.Email, fixt.Password, fixt.Active).
			WillReturnError(sql.ErrConnDone)

		err := s.repo.Save(s.ctx, fixt)

		s.Error(err)
	})
}
