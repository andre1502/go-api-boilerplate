package middleware

import (
	"fmt"
	"go-api-boilerplate/internal/response"
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/config"
	"go-api-boilerplate/module/redis"
	"go-api-boilerplate/module/token"
	"strings"

	"github.com/labstack/echo/v4"
)

type Middleware struct {
	Config   *config.Config
	Redis    *redis.RedisConnection
	Response *response.Response
}

func NewMiddleware(
	cfg *config.Config,
	rds *redis.RedisConnection,
	resp *response.Response,
) *Middleware {
	return &Middleware{
		Config:   cfg,
		Redis:    rds,
		Response: resp,
	}
}

func (m *Middleware) ExtractToken(c echo.Context) string {
	token := c.QueryParam("token")
	if !module.IsEmptyString(token) {
		return token
	}

	bearerToken := c.Request().Header.Get("Authorization")
	splitToken := strings.Split(bearerToken, " ")

	if len(splitToken) == 2 {
		return splitToken[1]
	}

	return ""
}

func (m *Middleware) GetJwtClaims(c echo.Context, jwtRedisKey string) (jwtClaims *token.JwtMapClaims, tkn string, err error) {
	requestTkn := m.ExtractToken(c)

	if module.IsEmptyString(requestTkn) {
		return nil, "", m.Response.Unauthorized(c)
	}

	jwtClaims, err = token.ParseJwtClaims(requestTkn, []byte(m.Config.JWT_SECRET_KEY))
	if err != nil {
		return nil, "", m.Response.Unauthorized(c)
	}

	redisKey := fmt.Sprintf(jwtRedisKey, jwtClaims.UID)

	tkn, err = m.Redis.GetCache(c.Request().Context(), redisKey)
	if err != nil || tkn != requestTkn {
		return nil, "", m.Response.Unauthorized(c)
	}

	return jwtClaims, tkn, nil
}
