package authentication

import (
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

type loginClaims struct {
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
	claims := &loginClaims{
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