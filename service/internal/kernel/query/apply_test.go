package query

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	mockDB, _, _ := sqlmock.New()
	db, _ := gorm.Open(postgres.New(postgres.Config{Conn: mockDB}), &gorm.Config{})
	return db
}

func TestApplyQuery_Branches(t *testing.T) {
	db := testDB(t)

	t.Run("nil query", func(t *testing.T) {
		if ApplyQuery(db, nil) == nil {
			t.Error("expected tx")
		}
	})

	t.Run("skips invalid sorts", func(t *testing.T) {
		tx := ApplyQuery(db, &Query{
			Sorts: []Sort{
				{Field: "created_at", Direction: "sideways"},
				{Field: "created_at; DROP TABLE x", Direction: Ascending},
				{Field: "created_at", Direction: Descending},
			},
		})
		if tx == nil {
			t.Error("expected tx")
		}
	})

	t.Run("skips unsafe filters", func(t *testing.T) {
		tx := ApplyQuery(db, &Query{
			Filters: []Filter{
				{Field: "x; DROP", Operator: Equal, Value: 1},
				{Field: "state", Operator: Operator("bogus"), Value: 1},
			},
		})
		if tx == nil {
			t.Error("expected tx")
		}
	})
}
