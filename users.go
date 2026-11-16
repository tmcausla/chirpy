package main
import (
	"log"
	"net/http"
	"encoding/json"
	"time"
	"github.com/google/uuid"
	"github.com/tmcausla/chirpy/internal/database"
	"github.com/tmcausla/chirpy/internal/auth"
)

type User struct {
	ID 		  	 uuid.UUID `json:"id"`
	CreatedAt 	 time.Time `json:"created_at"`
	UpdatedAt 	 time.Time `json:"updated_at"`
	Email 	  	 string    `json:"email"`
	Token	  	 string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type userLogin struct {
	Password 		 string `json:"password"`
	Email 	 		 string `json:"email"`
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	newUser := userLogin{}
	err := decoder.Decode(&newUser)
	if err != nil {
		log.Printf("error decoding request: %v\n", err)
        respondWithError(w, 400, "error decoding request")
        return
	}

	hashPass, err := auth.HashPassword(newUser.Password)
	if err != nil {
		log.Printf("%v\n", err)
		respondWithError(w, 500, "error hashing password")
		return
	}

	dbUser, err := cfg.dbQueries.CreateUser(req.Context(), database.CreateUserParams{
		Email: newUser.Email,
		HashedPassword: hashPass,
	})
	if err != nil {
		log.Printf("error creating user: %v\n", err)
		respondWithError(w, 500, "error creating user")
		return
	}

	respondWithJSON(w, 201, User{
		ID: dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email: dbUser.Email,
	})
}

func (cfg *apiConfig) login(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	loginUser := userLogin{}
	err := decoder.Decode(&loginUser)
	if err != nil {
		log.Printf("error logging in: %v\n", err)
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	dbUser, err := cfg.dbQueries.GetUserByEmail(req.Context(), loginUser.Email)
	if err != nil {
		log.Printf("error logging in: %v\n", err)
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	match, err := auth.CheckPasswordHash(loginUser.Password, dbUser.HashedPassword)
	if err != nil || !match {
		log.Printf("error logging in: %v\n", err)
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	newToken, err := auth.MakeJWT(dbUser.ID, cfg.tokenSecret, time.Hour)
	if err != nil {
		log.Printf("error logging in: %v\n", err)
		respondWithError(w, 500, "Internal Server Error")
		return
	}

	newRefreshToken, err := cfg.dbQueries.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
		Token: auth.MakeRefreshToken(),
		UserID: dbUser.ID,	
		ExpiresAt: time.Now().UTC().Add(60 * 24 * time.Hour),
	})
	if err != nil {
		log.Printf("error making refresh token: %v", err)
		respondWithError(w, 500, "error logging in")
		return
	}

	respondWithJSON(w, 200, User{
		ID: dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email: dbUser.Email,
		Token: newToken,
		RefreshToken: newRefreshToken.Token,
	})
}

func (cfg *apiConfig) updateUser(w http.ResponseWriter, req *http.Request) {
	authToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("error authenticating user: %v\n", err)
		respondWithError(w, 401, "401 Unauthorized")
		return
	}
	authID, err := auth.ValidateJWT(authToken, cfg.tokenSecret)
	if err != nil {
		log.Printf("error authenticating user: %v\n", err)
		respondWithError(w, 401, "401 Unauthorized")
		return
	}

	decoder := json.NewDecoder(req.Body)
	user := userLogin{}
	err = decoder.Decode(&user)
	if err != nil {
		log.Printf("error decoding request: %v\n", err)
        respondWithError(w, 400, "error decoding request")
        return
	}

	hashPass, err := auth.HashPassword(user.Password)
	if err != nil {
		log.Printf("%v\n", err)
		respondWithError(w, 500, "error hashing password")
		return
	}

	dbUser, err := cfg.dbQueries.UpdateUser(req.Context(), database.UpdateUserParams{
		ID: authID,
		Email: user.Email,
		HashedPassword: hashPass,
	})
	if err != nil {
		log.Printf("error updating user: %v\n", err)
		respondWithError(w, 500, "error updating user")
		return
	}

	respondWithJSON(w, 200, User{
		ID: dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email: dbUser.Email,
	})
}