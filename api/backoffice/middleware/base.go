package middleware

import (
	"go-api-boilerplate/internal/middleware"
)

type Middleware struct {
	*middleware.Middleware
}

func NewMiddleware(
	mdl *middleware.Middleware,
) *Middleware {
	return &Middleware{mdl}
}
