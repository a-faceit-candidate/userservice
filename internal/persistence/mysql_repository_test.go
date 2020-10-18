package persistence

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/a-faceit-candidate/userservice/internal/model"
	"github.com/go-playground/assert/v2"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func TestMysqlRepository_Create(t *testing.T) {
	// The test below tests the scenario that can't be tested by the acceptance test:
	// our service has generated same uuidV1 for two different users
	// we don't test the query itself: that's already tested in sqlbuilder, and testing it here would be fragile
	// and we don't test the fields being properly inserted in mysql
	// but we test the error coming from mysql instead, and how it's mapped to our logic.
	//
	// I hope this example serves as testing example here, filling hundreds of lines of tests here for a code challenge
	// seems like an unnecessary effort IMO.
	// If you've read until here. Thanks for reading.
	t.Run("duplicated row exception", func(t *testing.T) {
		mockedDB, mysqlMock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockedDB.Close()

		mysqlMock.ExpectExec("INSERT INTO user .*").WithArgs().
			// error code is defined here: https://dev.mysql.com/doc/refman/5.7/en/server-error-reference.html#error_er_dup_entry
			// we intentionally not use the constant we have
			WillReturnError(&mysql.MySQLError{Number: 1062})

		repo := NewMysqlRepository(mockedDB)
		err = repo.Create(context.Background(), &model.User{ID: "asdf"})
		assert.Equal(t, ErrConflict, err)
	})
}
