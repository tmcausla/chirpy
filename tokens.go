package main
import (
	"log"
	"time"
	"net/http"
	"github.com/tmcausla/chirpy/internal/auth"
)

func (cfg *apiConfig) refresh(w http.ResponseWriter, req *http.Request) {
	authToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("error authenticating user: %v\n", err)
		respondWithError(w, 401, "401 Unauthorized")
		return
	}

	userID, err := cfg.dbQueries.GetUserByRefreshToken(req.Context(), authToken)
	if err != nil {
		log.Printf("error fetching user: %v\n", err)
		respondWithError(w, 401, "401 Unauthorized")
		return
	}

	newAccessToken, err := auth.MakeJWT(userID, cfg.tokenSecret, time.Hour)
	if err != nil {
		log.Printf("error creating new token: %v\n", err)
		respondWithError(w, 500, "Internal Server Error")
		return
	}

	type tokenResp struct {
		Token string `json:"token"`
	}

	respondWithJSON(w, 200, tokenResp{
		Token: newAccessToken,
	})
}

func (cfg *apiConfig) revoke(w http.ResponseWriter, req *http.Request) {
	authToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("error authenticating user: %v\n", err)
		respondWithError(w, 401, "401 Unauthorized")
		return
	}

	_, err = cfg.dbQueries.RevokeToken(req.Context(), authToken)
	if err != nil {
		log.Printf("error revoking token: %v\n", err)
		respondWithError(w, 500, "Internal Server Error")
		return
	}

	w.WriteHeader(204)
}