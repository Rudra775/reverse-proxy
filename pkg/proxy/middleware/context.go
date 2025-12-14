package middleware

// ctxKey prevents collisions with other context keys
type ctxKey string

// RequestIDKey is used to store request ID in context
const RequestIDKey ctxKey = "request_id"
