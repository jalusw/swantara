package audit

import (
	"reflect"
	"strings"
	"time"
)

const RedactedValue = "[REDACTED]"

// Diff yields the captured change between two entities. A nil before captures
// the full after value (insert); a nil after captures the full before value
// (delete); otherwise only the fields that differ are captured (update).
func Diff(before, after any) map[string]any {
	switch {
	case before == nil && after == nil:
		return map[string]any{}
	case before == nil:
		return anyToMap(after)
	case after == nil:
		return anyToMap(before)
	default:
		oldValue, newValue := unwrap(reflect.ValueOf(before)), unwrap(reflect.ValueOf(after))
		if !oldValue.IsValid() || !newValue.IsValid() {
			return anyToMap(after)
		}
		return diffStruct(oldValue, newValue)
	}
}

func anyToMap(value any) map[string]any {
	after := unwrap(reflect.ValueOf(value))
	if !after.IsValid() {
		return map[string]any{}
	}
	return diffStruct(reflect.Zero(after.Type()), after)
}

func diffStruct(before, after reflect.Value) map[string]any {
	changes := map[string]any{}
	if before.Kind() != reflect.Struct || after.Kind() != reflect.Struct {
		return map[string]any{"value": after.Interface()}
	}

	typ := before.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name, skip := fieldName(field)
		if skip {
			continue
		}
		oldValue, newValue := before.Field(i), after.Field(i)
		if equal(oldValue, newValue) {
			continue
		}
		if field.Tag.Get("audit") == "redact" {
			changes[name] = RedactedValue
			continue
		}
		changes[name] = diffValue(oldValue, newValue)
	}

	return changes
}

func fieldName(field reflect.StructField) (string, bool) {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return "", true
	}
	if name, _, _ := strings.Cut(tag, ","); name != "" {
		return name, false
	}
	return field.Name, false
}

func equal(before, after reflect.Value) bool {
	oldValue, newValue := unwrap(before), unwrap(after)
	if !oldValue.IsValid() || !newValue.IsValid() {
		return oldValue.IsValid() == newValue.IsValid()
	}
	if oldValue.Type() != newValue.Type() {
		return false
	}
	if oldValue.Type() == reflect.TypeOf(time.Time{}) {
		return oldValue.Interface().(time.Time).Equal(newValue.Interface().(time.Time))
	}
	return reflect.DeepEqual(oldValue.Interface(), newValue.Interface())
}

func diffValue(before, after reflect.Value) any {
	oldValue, newValue := unwrap(before), unwrap(after)
	if !oldValue.IsValid() || !newValue.IsValid() || oldValue.Type() != newValue.Type() {
		return after.Interface()
	}
	if oldValue.Kind() == reflect.Struct && oldValue.Type() != reflect.TypeOf(time.Time{}) {
		return diffStruct(oldValue, newValue)
	}
	return newValue.Interface()
}

func unwrap(value reflect.Value) reflect.Value {
	for value.IsValid() && value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}
