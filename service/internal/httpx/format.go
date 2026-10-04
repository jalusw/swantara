// Package httpx format negotiation: opt-in XML/CSV with JSON default.
//
// JSON is always the default for both directions:
//   - Response: JSON unless the client explicitly opts in via ?format=xml|csv
//     or Accept: application/xml, text/csv AND the endpoint opted in.
//   - Request: JSON unless the client sends Content-Type: application/xml,
//     text/xml, text/csv AND the endpoint opted in.
//
// Endpoints opt in with WithFormats middleware (route level) or AllowFormats
// (inside the handler, before binding/responding). When no allowed formats
// are set, all three are accepted for backward compatibility.
package httpx

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Format is a negotiable payload format. JSON is the default.
type Format string

const (
	FormatJSON Format = "json"
	FormatXML  Format = "xml"
	FormatCSV  Format = "csv"
)

const (
	allowedResponseFormatsKey = "allowed_response_formats"
	allowedRequestFormatsKey  = "allowed_request_formats"
)

// WithFormats opts the route in to the given request + response formats.
// JSON is the default and does not need to be listed, but listing it
// explicitly is harmless. Example:
//
//	products.Get("/", httpx.WithFormats(httpx.FormatJSON, httpx.FormatCSV), guards.Guard("item", "view"), h.List)
func WithFormats(formats ...Format) fiber.Handler {
	normalized := normalizeFormats(formats)
	return func(c fiber.Ctx) error {
		c.Locals(allowedResponseFormatsKey, normalized)
		c.Locals(allowedRequestFormatsKey, normalized)
		return c.Next()
	}
}

// WithResponseFormats opts the route in to the given response formats only.
func WithResponseFormats(formats ...Format) fiber.Handler {
	normalized := normalizeFormats(formats)
	return func(c fiber.Ctx) error {
		c.Locals(allowedResponseFormatsKey, normalized)
		return c.Next()
	}
}

// WithRequestFormats opts the route in to the given request body formats only.
func WithRequestFormats(formats ...Format) fiber.Handler {
	normalized := normalizeFormats(formats)
	return func(c fiber.Ctx) error {
		c.Locals(allowedRequestFormatsKey, normalized)
		return c.Next()
	}
}

// AllowFormats opts the current handler in to the given request + response
// formats. Call before BindAndValidate / response helpers.
func AllowFormats(c fiber.Ctx, formats ...Format) {
	normalized := normalizeFormats(formats)
	c.Locals(allowedResponseFormatsKey, normalized)
	c.Locals(allowedRequestFormatsKey, normalized)
}

// AllowResponseFormats opts the current handler in to the given response formats.
func AllowResponseFormats(c fiber.Ctx, formats ...Format) {
	c.Locals(allowedResponseFormatsKey, normalizeFormats(formats))
}

// AllowRequestFormats opts the current handler in to the given request body formats.
func AllowRequestFormats(c fiber.Ctx, formats ...Format) {
	c.Locals(allowedRequestFormatsKey, normalizeFormats(formats))
}

func normalizeFormats(formats []Format) []Format {
	if len(formats) == 0 {
		return []Format{FormatJSON}
	}
	seen := make(map[Format]struct{}, len(formats))
	out := make([]Format, 0, len(formats))
	for _, f := range formats {
		n := Format(strings.ToLower(strings.TrimSpace(string(f))))
		switch n {
		case FormatJSON, FormatXML, FormatCSV:
			if _, ok := seen[n]; !ok {
				seen[n] = struct{}{}
				out = append(out, n)
			}
		}
	}
	if len(out) == 0 {
		return []Format{FormatJSON}
	}
	return out
}

func getAllowedFormats(c fiber.Ctx, key string) ([]Format, bool) {
	if c == nil {
		return nil, false
	}
	switch v := c.Locals(key).(type) {
	case []Format:
		return v, true
	case []string:
		out := make([]Format, 0, len(v))
		for _, s := range v {
			out = append(out, Format(strings.ToLower(strings.TrimSpace(s))))
		}
		return out, true
	default:
		return nil, false
	}
}

// IsResponseFormatAllowed reports whether f may be used for the response.
// When the endpoint never opted in, all formats are allowed (legacy).
func IsResponseFormatAllowed(c fiber.Ctx, f Format) bool {
	allowed, isSet := getAllowedFormats(c, allowedResponseFormatsKey)
	if !isSet {
		return f == FormatJSON || f == FormatXML || f == FormatCSV
	}
	for _, a := range allowed {
		if a == f {
			return true
		}
	}
	return false
}

// IsRequestFormatAllowed reports whether f may be used for the request body.
// When the endpoint never opted in, JSON/XML/CSV are allowed (legacy).
func IsRequestFormatAllowed(c fiber.Ctx, f Format) bool {
	allowed, isSet := getAllowedFormats(c, allowedRequestFormatsKey)
	if !isSet {
		return f == FormatJSON || f == FormatXML || f == FormatCSV
	}
	for _, a := range allowed {
		if a == f {
			return true
		}
	}
	return false
}

// explicitQueryFormat returns the ?format= override when present and valid.
func explicitQueryFormat(c fiber.Ctx) (Format, bool) {
	raw := strings.TrimSpace(c.Query("format"))
	if raw == "" {
		return "", false
	}
	switch f := Format(strings.ToLower(raw)); f {
	case FormatJSON, FormatXML, FormatCSV:
		return f, true
	default:
		return "", false
	}
}

// RequestedResponseFormat is the raw client preference: ?format= wins,
// then Accept header, defaulting to JSON.
func RequestedResponseFormat(c fiber.Ctx) Format {
	if f, ok := explicitQueryFormat(c); ok {
		return f
	}
	switch c.Accepts(fiber.MIMEApplicationJSON, fiber.MIMEApplicationXML, fiber.MIMETextXML, MIMETextCSV) {
	case fiber.MIMEApplicationXML, fiber.MIMETextXML:
		return FormatXML
	case MIMETextCSV:
		return FormatCSV
	default:
		return FormatJSON
	}
}

// NegotiatedResponseFormat intersects the client preference with the
// endpoint's opt-in. Disallowed formats fall back to JSON (default).
func NegotiatedResponseFormat(c fiber.Ctx) Format {
	requested := RequestedResponseFormat(c)
	if IsResponseFormatAllowed(c, requested) {
		return requested
	}
	return FormatJSON
}

// RequestFormat is the negotiated response format (allowed-aware).
// JSON is the default. Kept for backward compatibility.
func RequestFormat(c fiber.Ctx) Format {
	return NegotiatedResponseFormat(c)
}

// HasExplicitFormatQuery reports whether ?format= explicitly requests f.
func HasExplicitFormatQuery(c fiber.Ctx, f Format) bool {
	explicit, ok := explicitQueryFormat(c)
	return ok && explicit == f
}

// RequestBodyFormat detects the request body format from Content-Type.
// Returns "" for form/multipart (passthrough to Fiber) and other
// non-negotiable bodies. Empty Content-Type defaults to JSON.
func RequestBodyFormat(c fiber.Ctx) Format {
	ct := strings.ToLower(strings.TrimSpace(c.Get(fiber.HeaderContentType)))
	if i := strings.Index(ct, ";"); i >= 0 {
		ct = strings.TrimSpace(ct[:i])
	}
	if ct == "" {
		return FormatJSON
	}
	if strings.Contains(ct, "json") || strings.HasSuffix(ct, "+json") {
		return FormatJSON
	}
	if strings.Contains(ct, "xml") || strings.HasSuffix(ct, "+xml") {
		return FormatXML
	}
	if strings.Contains(ct, "csv") {
		return FormatCSV
	}
	return ""
}

func isFormContentType(c fiber.Ctx) bool {
	ct := strings.ToLower(strings.TrimSpace(c.Get(fiber.HeaderContentType)))
	if i := strings.Index(ct, ";"); i >= 0 {
		ct = strings.TrimSpace(ct[:i])
	}
	return ct == fiber.MIMEApplicationForm || ct == fiber.MIMEMultipartForm || strings.HasPrefix(ct, "multipart/")
}

// BindBody parses the request body according to Content-Type without
// validation: JSON via Fiber, XML via a json-tag-aware decoder, CSV via the
// local decoder, forms via Fiber. Unknown types fall back to Fiber's Body binder.
func BindBody(c fiber.Ctx, out any) error {
	if isFormContentType(c) {
		return c.Bind().Body(out)
	}
	switch RequestBodyFormat(c) {
	case FormatXML:
		if len(bytes.TrimSpace(c.Body())) == 0 {
			return fmt.Errorf("empty xml body")
		}
		return decodeXMLBody(c.Body(), out)
	case FormatCSV:
		if len(bytes.TrimSpace(c.Body())) == 0 {
			return fmt.Errorf("empty csv body")
		}
		return decodeCSVBody(c.Body(), out)
	case FormatJSON:
		if len(bytes.TrimSpace(c.Body())) == 0 {
			return c.Bind().Body(out)
		}
		return c.Bind().JSON(out)
	default:
		return c.Bind().Body(out)
	}
}

// decodeXMLBody unmarshals XML into out while honoring `json` tag names, so
// existing request structs (json-only) accept `<name>` as well as `<Name>`.
// Nested objects and repeated elements (slices) are supported via a
// map-roundtrip through encoding/json. Falls back to encoding/xml for
// structs using explicit `xml` tags.
func decodeXMLBody(body []byte, out any) error {
	if err := decodeXMLViaMap(body, out); err == nil {
		return nil
	}
	return xml.Unmarshal(body, out)
}

func decodeXMLViaMap(body []byte, out any) error {
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("xml destination must be a non-nil pointer")
	}
	dec := xml.NewDecoder(bytes.NewReader(body))
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		if start, ok := tok.(xml.StartElement); ok {
			val, err := parseXMLElement(dec, start)
			if err != nil {
				return err
			}
			var payload []byte
			if m, ok := val.(map[string]any); ok {
				payload, err = json.Marshal(m)
			} else if s, ok := val.(string); ok {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("empty xml body")
				}
				return fmt.Errorf("xml body must contain named fields")
			} else {
				payload, err = json.Marshal(val)
			}
			if err != nil {
				return err
			}
			return json.Unmarshal(payload, out)
		}
	}
}

func parseXMLElement(dec *xml.Decoder, start xml.StartElement) (any, error) {
	children := make(map[string]any)
	var text strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			child, err := parseXMLElement(dec, t)
			if err != nil {
				return nil, err
			}
			name := t.Name.Local
			if existing, ok := children[name]; ok {
				switch ex := existing.(type) {
				case []any:
					children[name] = append(ex, child)
				default:
					children[name] = []any{ex, child}
				}
			} else {
				children[name] = child
			}
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				if len(children) > 0 {
					return children, nil
				}
				return strings.TrimSpace(text.String()), nil
			}
			return nil, fmt.Errorf("mismatched xml end tag %q", t.Name.Local)
		case xml.CharData:
			text.Write([]byte(t))
		}
	}
}

// decodeCSVBody maps a CSV document (header + rows) onto out, which must be
// a pointer to a struct (single row) or a pointer to a slice/array.
// Columns match struct `json` tag names, mirroring ExportCSV.
func decodeCSVBody(body []byte, out any) error {
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("csv destination must be a non-nil pointer")
	}
	elem := rv.Elem()
	switch elem.Kind() {
	case reflect.Struct:
		rows, err := decodeCSVRows(body, elem.Type())
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return fmt.Errorf("csv body must contain a header and one data row")
		}
		elem.Set(rows[0])
		return nil
	case reflect.Slice, reflect.Array:
		itemType := elem.Type().Elem()
		if itemType.Kind() == reflect.Pointer {
			itemType = itemType.Elem()
		}
		if itemType.Kind() != reflect.Struct {
			return fmt.Errorf("csv slice elements must be structs")
		}
		rows, err := decodeCSVRows(body, itemType)
		if err != nil {
			return err
		}
		slice := reflect.MakeSlice(elem.Type(), len(rows), len(rows))
		for i, row := range rows {
			dst := slice.Index(i)
			if dst.Kind() == reflect.Pointer {
				ptr := reflect.New(itemType)
				ptr.Elem().Set(row)
				dst.Set(ptr)
			} else {
				dst.Set(row)
			}
		}
		elem.Set(slice)
		return nil
	default:
		return fmt.Errorf("csv destination must be a struct or slice pointer")
	}
}

func decodeCSVRows(body []byte, itemType reflect.Type) ([]reflect.Value, error) {
	reader := csv.NewReader(bytes.NewReader(body))
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 1 {
		return nil, fmt.Errorf("csv body must contain a header row")
	}
	header := records[0]
	colIndex := make(map[string]int, len(header))
	for i, h := range header {
		name := strings.TrimSpace(strings.ToLower(h))
		if name != "" {
			if _, exists := colIndex[name]; !exists {
				colIndex[name] = i
			}
		}
	}
	fieldByCol := make(map[int]int)
	for i := 0; i < itemType.NumField(); i++ {
		field := itemType.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name := strings.ToLower(csvFieldName(field))
		if name == "" || !isCSVScalar(field.Type) {
			continue
		}
		if ci, ok := colIndex[name]; ok {
			fieldByCol[ci] = i
		}
	}
	var rows []reflect.Value
	for _, record := range records[1:] {
		if len(record) == 0 {
			continue
		}
		empty := true
		for _, cell := range record {
			if strings.TrimSpace(cell) != "" {
				empty = false
				break
			}
		}
		if empty {
			continue
		}
		row := reflect.New(itemType).Elem()
		for ci, fi := range fieldByCol {
			var cell string
			if ci < len(record) {
				cell = strings.TrimSpace(record[ci])
			}
			if err := setCSVCell(row.Field(fi), cell); err != nil {
				return nil, err
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func setCSVCell(field reflect.Value, cell string) error {
	if !field.CanSet() {
		return nil
	}
	t := field.Type()
	if t.Kind() == reflect.Pointer {
		if cell == "" {
			return nil
		}
		ptr := reflect.New(t.Elem())
		if err := setCSVCell(ptr.Elem(), cell); err != nil {
			return err
		}
		field.Set(ptr)
		return nil
	}
	switch t.Kind() {
	case reflect.String:
		field.SetString(cell)
		return nil
	case reflect.Bool:
		if cell == "" {
			return nil
		}
		b, err := strconv.ParseBool(cell)
		if err != nil {
			return fmt.Errorf("invalid bool %q", cell)
		}
		field.SetBool(b)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if cell == "" {
			return nil
		}
		n, err := strconv.ParseInt(cell, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer %q", cell)
		}
		field.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if cell == "" {
			return nil
		}
		n, err := strconv.ParseUint(cell, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unsigned integer %q", cell)
		}
		field.SetUint(n)
		return nil
	case reflect.Float32, reflect.Float64:
		if cell == "" {
			return nil
		}
		f, err := strconv.ParseFloat(cell, 64)
		if err != nil {
			return fmt.Errorf("invalid number %q", cell)
		}
		field.SetFloat(f)
		return nil
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			if cell == "" {
				return nil
			}
			parsed, err := time.Parse(time.RFC3339, cell)
			if err != nil {
				if d, derr := time.Parse("2006-01-02", cell); derr == nil {
					field.Set(reflect.ValueOf(d))
					return nil
				}
				return fmt.Errorf("invalid time %q, want RFC3339", cell)
			}
			field.Set(reflect.ValueOf(parsed))
			return nil
		}
	}
	return fmt.Errorf("unsupported csv field type %s", t.String())
}
