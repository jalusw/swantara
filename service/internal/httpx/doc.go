// Package httpx is the shared HTTP layer built on Fiber v3. It creates the app
// with middleware, defines the standard response envelope, request binding and
// validation, query parsing, auth guards, tenant resolution, idempotency, rate
// limiting, and health and swagger wiring.
//
// Format negotiation is opt-in with JSON as the default: responses honor
// ?format=json|xml|csv and the Accept header, requests honor Content-Type
// (application/json, application/xml, text/xml, text/csv). Endpoints opt in
// via WithFormats middleware or AllowFormats inside the handler; without
// opt-in all formats are accepted for backward compatibility. Use
// WriteListResponse for lists to get CSV export with a custom filename, and
// BindAndValidate (or BindAndValidateWithFormats) for body parsing.
package httpx
