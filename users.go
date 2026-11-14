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
	ID 		  uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email 	  string 	`json:"email"`
}

type userLogin struct {
	Password string `json:"password"`
	Email 	 string `json:"email"`
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

	user := User{
		ID: dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email: dbUser.Email,
	}
	respondWithJSON(w, 201, user)
}

func (cfg *apiConfig) login(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	loginUser := userLogin{}
	err := decoder.Decode(&loginUser)
	if err != nil {
		log.Println("error logging in")
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	dbUser, err := cfg.dbQueries.GetUserByEmail(req.Context(), loginUser.Email)
	if err != nil {
		log.Println("error logging in")
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	match, err := auth.CheckPasswordHash(loginUser.Password, dbUser.HashedPassword)
	if err != nil || !match {
		log.Println("error logging in")
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	user := User{
		ID: dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email: dbUser.Email,
	}
	respondWithJSON(w, 200, user)
}