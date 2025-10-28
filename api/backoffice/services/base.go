package services

import (
	"context"
	"fmt"
	"go-api-boilerplate/internal/constant"
	"go-api-boilerplate/internal/exception"
	"go-api-boilerplate/internal/services"
	"go-api-boilerplate/internal/status_code"
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/token"
	"time"

	"github.com/labstack/echo/v4"
)

type Service struct {
	*services.Service
}

func NewService(
	svc *services.Service,
) *Service {
	return &Service{
		Service: svc,
	}
}

func (s *Service) ClearCacheBOUserInfo(ctx context.Context, userId int) {
	key := fmt.Sprintf(constant.BO_USER_INFO, userId)

	s.Redis.DelCache(ctx, key)
}

func (s *Service) ClearCacheBOUser(ctx context.Context, userId int) {
	keys := []string{
		fmt.Sprintf(constant.BO_USER_JWT, userId),
		fmt.Sprintf(constant.BO_USER_INFO, userId),
	}

	s.Redis.DelCache(ctx, keys...)
}

func (s *Service) GenerateJwt(c echo.Context, requestUri string, userID int, account string, roleID int) (string, *token.JwtMapClaims, error) {
	tokenLifeSpan := module.StrToIntDefault(s.Config.BACKOFFICE_JWT_EXPIRE_MINUTES, 2880)
	customClaims := token.CustomClaims{
		UID:     userID,
		Account: &account,
		RoleID:  &roleID,
	}

	tkn, claims, err := token.GenerateJwt(requestUri, customClaims, tokenLifeSpan, []byte(s.Config.JWT_SECRET_KEY))
	if err != nil {
		return "", nil, exception.Ex.Errors(status_code.UNAUTHORIZED_ERROR_CODE, status_code.UNAUTHORIZED_ERROR_MESSAGE, err)
	}

	redisKey := fmt.Sprintf(constant.BO_USER_JWT, customClaims.UID)

	if err = s.Redis.SetCache(c.Request().Context(), redisKey, tkn, time.Duration(tokenLifeSpan)*time.Minute); err != nil {
		return "", nil, exception.Ex.MappingErrorRedis(err)
	}

	return tkn, claims, nil
}
