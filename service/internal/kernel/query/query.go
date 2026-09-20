package query

import (
	"regexp"

	"gorm.io/gorm"
)

var safeFieldRe = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

const (
	Ascending             = "ASC"
	Descending            = "DESC"
	Equal        Operator = "="
	NotEqual     Operator = "!="
	Greater      Operator = ">"
	GreaterEqual Operator = ">="
	Less         Operator = "<"
	LessEqual    Operator = "<="
	Like         Operator = "LIKE"
	In           Operator = "IN"
	NotIn        Operator = "NOT_IN"
	IsNull       Operator = "IS NULL"
)

type Operator string
type Direction string

type Query struct {
	Pagination *Pagination
	Sorts      []Sort
	Filters    []Filter
}

type Page[T any] struct {
	Items []*T
	Count int64
}

type Pagination struct {
	Page int
	Size int
}

type Filter struct {
	Field    string
	Operator Operator
	Value    any
}

type Sort struct {
	Field     string
	Direction Direction
}

func ApplyQuery(tx *gorm.DB, query *Query) *gorm.DB {
	if query == nil {
		return tx
	}

	tx = ApplyFilters(tx, query.Filters)

	for _, v := range query.Sorts {
		if v.Direction != Ascending && v.Direction != Descending {
			continue
		}
		if !safeFieldRe.MatchString(v.Field) {
			continue
		}
		tx = tx.Order(v.Field + " " + string(v.Direction))
	}

	if query.Pagination != nil {
		tx = tx.Offset((query.Pagination.Page - 1) * query.Pagination.Size).Limit(query.Pagination.Size)
	}

	return tx
}

func ApplyFilters(tx *gorm.DB, filters []Filter) *gorm.DB {
	for _, v := range filters {
		if !safeFieldRe.MatchString(v.Field) {
			continue
		}
		switch v.Operator {
		case Equal:
			tx = tx.Where(v.Field+" = ?", v.Value)
		case NotEqual:
			tx = tx.Where(v.Field+" != ?", v.Value)
		case Greater:
			tx = tx.Where(v.Field+" > ?", v.Value)
		case GreaterEqual:
			tx = tx.Where(v.Field+" >= ?", v.Value)
		case Less:
			tx = tx.Where(v.Field+" < ?", v.Value)
		case LessEqual:
			tx = tx.Where(v.Field+" <= ?", v.Value)
		case Like:
			tx = tx.Where(v.Field+" LIKE ?", v.Value)
		case In:
			tx = tx.Where(v.Field+" IN ?", v.Value)
		case NotIn:
			tx = tx.Where(v.Field+" NOT IN ?", v.Value)
		case IsNull:
			tx = tx.Where(v.Field + " IS NULL")
		}
	}

	return tx
}
