package auth 
import (
	"log"
	"strings"
	"errors"
	"net/http"
)

func GetAPIKey(headers http.Header) (string, error) {
	apiHeader := headers.Get("Authorization")
	if apiHeader == "" {
		log.Println("missing authorization header")
		return apiHeader, errors.New("missing authorization header")
	}

	arr := strings.Split(apiHeader, " ")
	if len(arr) == 2 && arr[0] == "ApiKey" {
		return arr[1], nil
	}

	log.Println("auth header is malformed")
	return "", errors.New("auth header is malformed")
}