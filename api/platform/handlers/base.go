package handlers

import (
	"go-api-boilerplate/internal/handlers"
)

type Handler struct {
	*handlers.Handler
}

func NewHandler(
	hdlr *handlers.Handler,
) *Handler {
	return &Handler{hdlr}
}
