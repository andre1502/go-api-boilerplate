package handlers

import (
	"go-api-boilerplate/internal/constant"
	"go-api-boilerplate/internal/exception"
	"go-api-boilerplate/internal/handlers"
	"go-api-boilerplate/internal/status_code"
	"go-api-boilerplate/module/token"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	*handlers.Handler
}

func NewHandler(
	hdlr *handlers.Handler,
) *Handler {
	return &Handler{hdlr}
}

func (h *Handler) GetCurrentUser(c echo.Context) (*token.CustomClaims, error) {
	ex := exception.Ex.Errors(status_code.UNAUTHORIZED_ERROR_CODE, status_code.UNAUTHORIZED_ERROR_MESSAGE, nil)

	userID, ok := c.Get(constant.BO_USER_ID).(int)
	if !ok {
		return nil, ex
	}

	account, ok := c.Get(constant.BO_USER_ACCOUNT).(string)
	if !ok {
		return nil, ex
	}

	roleId, ok := c.Get(constant.BO_ROLE_ID).(int)
	if !ok {
		return nil, ex
	}

	return &token.CustomClaims{
		UID:     userID,
		Account: &account,
		RoleID:  &roleId,
	}, nil
}
