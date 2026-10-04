package httpx

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

const MIMETextCSV = "text/csv; charset=utf-8"

func ExportCSV(c fiber.Ctx, status int, filename string, items any) error {
	if items == nil {
		return fmt.Errorf("csv export requires a slice, got nil")
	}
	itemsType := reflect.TypeOf(items)
	if itemsType.Kind() != reflect.Slice && itemsType.Kind() != reflect.Array {
		return fmt.Errorf("csv export requires a slice, got %s", itemsType.Kind())
	}
	elemType := itemsType.Elem()
	if elemType.Kind() == reflect.Pointer {
		elemType = elemType.Elem()
	}
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("csv export requires a slice of structs")
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writeCSVHeader(writer, elemType); err != nil {
		return err
	}
	if err := writeCSVRows(writer, items, elemType); err != nil {
		return err
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return err
	}

	c.Attachment(filename)
	c.Set(fiber.HeaderContentType, MIMETextCSV)
	return c.Status(status).SendString(buf.String())
}

type csvColumn struct {
	index int
}

func csvColumns(t reflect.Type) ([]string, []csvColumn) {
	var names []string
	var columns []csvColumn
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name := csvFieldName(field)
		if name == "" {
			continue
		}
		if !isCSVScalar(field.Type) {
			continue
		}
		names = append(names, name)
		columns = append(columns, csvColumn{index: i})
	}
	return names, columns
}

func isCSVScalar(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	case reflect.Pointer:
		return isCSVScalar(t.Elem())
	case reflect.Struct:
		return t == reflect.TypeOf(time.Time{})
	default:
		return false
	}
}

func csvFieldName(field reflect.StructField) string {
	if tag := field.Tag.Get("json"); tag != "" {
		if name, _, _ := strings.Cut(tag, ","); name != "" {
			return name
		}
	}
	return ""
}

func writeCSVHeader(writer *csv.Writer, t reflect.Type) error {
	names, _ := csvColumns(t)
	return writer.Write(names)
}

func writeCSVRows(writer *csv.Writer, items any, elemType reflect.Type) error {
	value := reflect.ValueOf(items)
	_, columns := csvColumns(elemType)
	for i := 0; i < value.Len(); i++ {
		item := value.Index(i)
		if item.Kind() == reflect.Pointer {
			if item.IsNil() {
				if err := writer.Write(make([]string, len(columns))); err != nil {
					return err
				}
				continue
			}
			item = item.Elem()
		}
		row := make([]string, len(columns))
		for j, column := range columns {
			row[j] = csvCellString(item.Field(column.index))
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func csvCellString(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Pointer:
		if v.IsNil() {
			return ""
		}
		return csvCellString(v.Elem())
	case reflect.Struct:
		if t, ok := v.Interface().(time.Time); ok {
			return t.Format(time.RFC3339)
		}
	}
	return ""
}
