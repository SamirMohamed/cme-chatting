package authentication

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

type LoginClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type Jwt struct {
	jwtSigningKey []byte
}

func NewJwtAuthenticator() *Jwt {
	return &Jwt{
		jwtSigningKey: []byte(os.Getenv("JWT_SIGNING_KEY")),
	}
}

func (j *Jwt) GenerateJWT(username string) (string, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &LoginClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.jwtSigningKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (j *Jwt) VerifyJWT(tokenString string) (*LoginClaims, error) {
	claims := &LoginClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return j.jwtSigningKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, fmt.Errorf("invalid signature")
		}
		return nil, fmt.Errorf("invalid token")
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
