package helper

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

type Int64Array []int64

func (a Int64Array) Value() (driver.Value, error) {
	parts := make([]string, len(a))
	for i, value := range a {
		parts[i] = strconv.FormatInt(value, 10)
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}

func (a *Int64Array) Scan(value any) error {
	if value == nil {
		*a = nil
		return nil
	}
	raw, ok := value.(string)
	if !ok {
		return fmt.Errorf("cannot scan %T into Int64Array", value)
	}
	raw = strings.Trim(raw, "{}")
	if raw == "" {
		*a = Int64Array{}
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make(Int64Array, len(parts))
	for i, part := range parts {
		parsed, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return err
		}
		out[i] = parsed
	}
	*a = out
	return nil
}
