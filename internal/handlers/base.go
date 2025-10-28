package handlers

import (
	"go-api-boilerplate/internal/constant"
	"go-api-boilerplate/internal/exception"
	"go-api-boilerplate/internal/response"
	"go-api-boilerplate/internal/status_code"
	"go-api-boilerplate/internal/validation"
	"go-api-boilerplate/module/token"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	Validation *validation.Validation
	Response   *response.Response
}

func NewHandler(
	vld *validation.Validation,
	resp *response.Response,
) *Handler {
	return &Handler{
		Validation: vld,
		Response:   resp,
	}
}

func (h *Handler) GetRequestIPAddress(c echo.Context) string {
	ipAddress, ok := c.Get(constant.IP_ADDRESS).(string)
	if !ok {
		return ""
	}

	return ipAddress
}

func (h *Handler) GetCurrentUser(c echo.Context) (*token.CustomClaims, error) {
	ex := exception.Ex.Errors(status_code.UNAUTHORIZED_ERROR_CODE, status_code.UNAUTHORIZED_ERROR_MESSAGE, nil)

	userID, ok := c.Get(constant.USER_ID).(int)
	if !ok {
		return nil, ex
	}

	userGUID, ok := c.Get(constant.USER_GUID).(int)
	if !ok {
		return nil, ex
	}

	return &token.CustomClaims{
		UID:  userID,
		GUID: &userGUID,
	}, nil
}
