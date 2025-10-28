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
)

type Service struct {
	*services.Service
}

func NewService(svc *services.Service) *Service {
	return &Service{svc}
}

func (s *Service) GenerateJwt(ctx context.Context, requestUri string, userID int, userGUID int) (tkn string, claims *token.JwtMapClaims, err error) {
	tokenLifeSpan := module.StrToIntDefault(s.Config.PLATFORM_JWT_EXPIRE_MINUTES, 4320)
	customClaims := token.CustomClaims{
		UID:  userID,
		GUID: &userGUID,
	}

	tkn, claims, err = token.GenerateJwt(requestUri, customClaims, tokenLifeSpan, []byte(s.Config.JWT_SECRET_KEY))
	if err != nil {
		return "", nil, exception.Ex.Errors(status_code.UNAUTHORIZED_ERROR_CODE, status_code.UNAUTHORIZED_ERROR_MESSAGE, err)
	}

	redisKey := fmt.Sprintf(constant.USER_JWT, customClaims.UID)

	if err = s.Redis.SetCache(ctx, redisKey, tkn, time.Duration(tokenLifeSpan)*time.Minute); err != nil {
		return "", nil, exception.Ex.MappingErrorRedis(err)
	}

	return tkn, claims, nil
}
