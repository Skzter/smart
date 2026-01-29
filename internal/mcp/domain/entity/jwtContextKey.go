package entity

// JwtContextKey is used as a unique key in the request's context.Context
// to store the JWT token for the current HTTP request. The type is an empty
// struct so context keys do not collide with other entries.
type JwtContextKey struct{}
