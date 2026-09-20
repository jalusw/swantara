package query

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type testEntity struct {
	ID   uint64 `gorm:"primaryKey"`
	Name string
}

func TestApplyQuery_Pagination(t *testing.T) {
	db, mock := NewMockDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "test_entities" LIMIT $1 OFFSET $2`)).
		WithArgs(5, 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	db = ApplyQuery(db, &Query{
		Pagination: &Pagination{Page: 3, Size: 5},
	})
	var entities []testEntity
	db.Find(&entities)

	AssertDBMockDone(t, mock)
}

func TestApplyQuery_Sort(t *testing.T) {
	db, mock := NewMockDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "test_entities" ORDER BY name ASC,created_at DESC`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	db = ApplyQuery(db, &Query{
		Sorts: []Sort{
			{Field: "name", Direction: Ascending},
			{Field: "created_at", Direction: Descending},
		},
	})
	var entities []testEntity
	db.Find(&entities)

	AssertDBMockDone(t, mock)
}

func TestApplyQuery_NoPagination(t *testing.T) {
	db, mock := NewMockDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "test_entities"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	db = ApplyQuery(db, &Query{})
	var entities []testEntity
	db.Find(&entities)

	AssertDBMockDone(t, mock)
}

func TestApplyFilters_EqualToIn(t *testing.T) {
	db, mock := NewMockDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "test_entities" WHERE name = $1 AND name != $2 AND name LIKE $3 AND name IN ($4,$5) AND name NOT IN ($6,$7)`)).
		WithArgs("val", "excluded", "%prefix%", "a", "b", "c", "d").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	db = ApplyFilters(db, []Filter{
		{Field: "name", Operator: Equal, Value: "val"},
		{Field: "name", Operator: NotEqual, Value: "excluded"},
		{Field: "name", Operator: Like, Value: "%prefix%"},
		{Field: "name", Operator: In, Value: []string{"a", "b"}},
		{Field: "name", Operator: NotIn, Value: []string{"c", "d"}},
	})
	var entities []testEntity
	db.Find(&entities)

	AssertDBMockDone(t, mock)
}

func TestApplyFilters_ComparisonOperators(t *testing.T) {
	db, mock := NewMockDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "test_entities" WHERE id > $1 AND id >= $2 AND id < $3 AND id <= $4`)).
		WithArgs(10, 5, 100, 50).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	db = ApplyFilters(db, []Filter{
		{Field: "id", Operator: Greater, Value: 10},
		{Field: "id", Operator: GreaterEqual, Value: 5},
		{Field: "id", Operator: Less, Value: 100},
		{Field: "id", Operator: LessEqual, Value: 50},
	})
	var entities []testEntity
	db.Find(&entities)

	AssertDBMockDone(t, mock)
}

func TestApplyFilters_IsNull(t *testing.T) {
	db, mock := NewMockDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "test_entities" WHERE organization_id IS NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	db = ApplyFilters(db, []Filter{
		{Field: "organization_id", Operator: IsNull},
	})
	var entities []testEntity
	db.Find(&entities)

	AssertDBMockDone(t, mock)
}

func TestApplyFilters_UnsupportedOperatorIgnored(t *testing.T) {
	db, mock := NewMockDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "test_entities"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	db = ApplyFilters(db, []Filter{
		{Field: "name", Operator: "BOGUS", Value: "x"},
	})
	var entities []testEntity
	db.Find(&entities)

	AssertDBMockDone(t, mock)
}

func TestDAOQuery_WithContext(t *testing.T) {
	db, mock := NewMockDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "test_entities"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	db = db.WithContext(context.Background())
	db = ApplyQuery(db, &Query{})
	var entities []testEntity
	db.Find(&entities)

	AssertDBMockDone(t, mock)
}

func TestDAOQuery_NilQuery(t *testing.T) {
	db, mock := NewMockDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "test_entities"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	db = ApplyQuery(db, nil)
	var entities []testEntity
	db.Find(&entities)

	AssertDBMockDone(t, mock)
}

func ExampleApplyQuery() {
	db, _ := NewMockDB(&testing.T{})

	_ = ApplyQuery(db, &Query{
		Filters: []Filter{
			{Field: "name", Operator: Equal, Value: "swantara"},
		},
		Sorts: []Sort{
			{Field: "created_at", Direction: Descending},
		},
	})
	// Output:
}

func ExamplePage() {
	_ = Page[testEntity]{Count: 1}
	// Output:
}
