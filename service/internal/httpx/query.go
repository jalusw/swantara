package httpx

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

const (
	defaultQueryPage = 1
	defaultQuerySize = 20
	maxQuerySize     = 100
)

var filterOperators = map[string]query.Operator{
	"eq":     query.Equal,
	"neq":    query.NotEqual,
	"gt":     query.Greater,
	"gte":    query.GreaterEqual,
	"lt":     query.Less,
	"lte":    query.LessEqual,
	"like":   query.Like,
	"in":     query.In,
	"not_in": query.NotIn,
}

func ParseQueryParams(c fiber.Ctx, allowlist map[string]struct{}) (*query.Query, error) {
	q := &query.Query{
		Pagination: &query.Pagination{Page: defaultQueryPage, Size: defaultQuerySize},
	}

	if raw := c.Query("page"); raw != "" {
		page, err := strconv.Atoi(raw)
		if err != nil || page < 1 {
			return nil, fmt.Errorf("invalid page: %q", raw)
		}
		q.Pagination.Page = page
	}

	if raw := c.Query("size"); raw != "" {
		size, err := strconv.Atoi(raw)
		if err != nil || size < 1 {
			return nil, fmt.Errorf("invalid size: %q", raw)
		}
		if size > maxQuerySize {
			size = maxQuerySize
		}
		q.Pagination.Size = size
	}

	if raw := c.Query("sort"); raw != "" {
		for token := range strings.SplitSeq(raw, ",") {
			field, direction, err := parseSortToken(token, allowlist)
			if err != nil {
				return nil, err
			}
			q.Sorts = append(q.Sorts, query.Sort{Field: field, Direction: direction})
		}
	}

	for _, raw := range c.RequestCtx().QueryArgs().PeekMulti("filter") {
		filter, err := parseFilterToken(string(raw), allowlist)
		if err != nil {
			return nil, err
		}
		q.Filters = append(q.Filters, filter)
	}

	return q, nil
}

func parseSortToken(token string, allowlist map[string]struct{}) (string, query.Direction, error) {
	field, direction, found := strings.Cut(token, ":")
	if !found {
		direction = "asc"
	}
	if _, ok := allowlist[field]; !ok {
		return "", "", fmt.Errorf("sort field %q is not allowed", field)
	}
	switch direction {
	case "asc":
		return field, query.Ascending, nil
	case "desc":
		return field, query.Descending, nil
	default:
		return "", "", fmt.Errorf("invalid sort direction %q", direction)
	}
}

func parseFilterToken(raw string, allowlist map[string]struct{}) (query.Filter, error) {
	field, rest, found := strings.Cut(raw, ":")
	if !found {
		return query.Filter{}, fmt.Errorf("invalid filter %q, expected field:op:value", raw)
	}
	if _, ok := allowlist[field]; !ok {
		return query.Filter{}, fmt.Errorf("filter field %q is not allowed", field)
	}
	op, value, found := strings.Cut(rest, ":")
	if !found {
		return query.Filter{}, fmt.Errorf("invalid filter %q, expected field:op:value", raw)
	}
	operator, ok := filterOperators[op]
	if !ok {
		return query.Filter{}, fmt.Errorf("invalid filter operator %q", op)
	}
	if operator == query.In || operator == query.NotIn {
		return query.Filter{Field: field, Operator: operator, Value: strings.Split(value, ",")}, nil
	}
	return query.Filter{Field: field, Operator: operator, Value: value}, nil
}

func BuildListMeta(q *query.Query, total int64) Meta {
	meta := Meta{}

	if q.Pagination != nil {
		size := q.Pagination.Size
		pages := int(total) / size
		if int(total)%size != 0 {
			pages++
		}
		meta.Pagination = &PaginationMeta{
			Page:       q.Pagination.Page,
			PerPage:    size,
			Total:      int(total),
			TotalPages: pages,
		}
	}

	if len(q.Sorts) > 0 {
		meta.Sort = &SortMeta{
			Field:     q.Sorts[0].Field,
			Direction: strings.ToLower(string(q.Sorts[0].Direction)),
		}
	}

	if len(q.Filters) > 0 {
		filters := make(FilterMeta, len(q.Filters))
		for _, f := range q.Filters {
			filters[f.Field] = filterValueString(f)
		}
		meta.Filter = filters
	}

	return meta
}

func filterValueString(f query.Filter) string {
	switch v := f.Value.(type) {
	case string:
		return v
	case []string:
		return strings.Join(v, ",")
	default:
		return fmt.Sprint(v)
	}
}
