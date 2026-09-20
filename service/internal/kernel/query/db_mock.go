package query

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	mockDB, mock, _ := sqlmock.New()
	db, _ := gorm.Open(postgres.New(postgres.Config{Conn: mockDB}), &gorm.Config{})
	return db, mock
}

func AssertDBMockDone(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
