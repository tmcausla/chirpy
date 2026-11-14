package auth
import (
	"log"
	"github.com/alexedwards/argon2id"
)

func HashPassword(pswd string) (string, error) {
	hash, err := argon2id.CreateHash(pswd, argon2id.DefaultParams)
	if err != nil {
		log.Printf("error hashing password: %v\n", err)
		return "", err
	}

	return hash, nil
}

func CheckPasswordHash(pswd, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(pswd, hash)
	if err != nil {
		log.Printf("error comparing password to hash: %v\n", err)
		return false, err
	}

	return match, nil
}