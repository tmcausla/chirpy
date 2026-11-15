package auth
import (
	"log"
	"time"
	"errors"
	"net/http"
	"strings"
	"crypto/rand"
	"encoding/hex"
	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
)

func MakeRefreshToken() string {
	randData := make([]byte, 32)
	_, err := rand.Read(randData)
	if err != nil {
		log.Printf("error creating refresh token: %v", err)
		return ""
	}
	hexStr := hex.EncodeToString(randData)

	return hexStr
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	tokenClaims := jwt.RegisteredClaims{
		Issuer: "chirpy-access",
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject: userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)

	newToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		log.Printf("error signing token: %v\n", err)
		return "", err
	}
	return newToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	tokenClaims := jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenString, &tokenClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		log.Printf("error parsing token: %v\n", err)
		return uuid.Nil, err
	}

	id, err := uuid.Parse(tokenClaims.Subject)
	if err != nil {
		log.Printf("invalid UUID: %v\n", err)
		return uuid.Nil, err
	}
	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		log.Println("missing authorization header")
		return authHeader, errors.New("missing authorization header")
	}

	arr := strings.Split(authHeader, " ")
	if len(arr) == 2 && arr[0] == "Bearer" {
		return arr[1], nil
	}

	log.Println("auth header is malformed")
	return "", errors.New("auth header is malformed")
}