package token

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type CustomClaims struct {
	UID     int     `json:"uid"`
	GUID    *int    `json:"gid,omitempty"`
	Account *string `json:"account,omitempty"`
	RoleID  *int    `json:"role_id,omitempty"`
}

type JwtMapClaims struct {
	CustomClaims
	jwt.RegisteredClaims
}

func GenerateJwt(requestUri string, customClaims CustomClaims, tokenLifeSpan int, secretKey []byte) (tkn string, claims *JwtMapClaims, err error) {
	now := time.Now()

	claims = &JwtMapClaims{
		CustomClaims: customClaims,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    requestUri,
			Subject:   strconv.Itoa(int(customClaims.UID)),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(tokenLifeSpan) * time.Minute)),
			NotBefore: jwt.NewNumericDate(now.Add(time.Duration(-10) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tkn, err = token.SignedString(secretKey)
	if err != nil {
		return "", nil, jwt.ErrTokenSignatureInvalid
	}

	return tkn, claims, nil
}

func ParseJwtClaims(tokenString string, secretKey []byte) (*JwtMapClaims, error) {
	claims := &JwtMapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", jwt.ErrTokenMalformed
		}

		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenUnverifiable
	}

	return claims, nil
}
