package lib

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TODO read from env variable
const secret = "test-secret"

func ParseToken(accessToken string) (int64, error) {
	token, err := jwt.ParseWithClaims(accessToken, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("token claims are not of type *tokenClaims")
	}
	if int64(claims["exp"].(float64)) < time.Now().Unix() {
		return 0, errors.New("token has expired")
	}
	return int64(claims["uid"].(float64)), nil
}
