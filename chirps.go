package main
import (
	"log"
	"net/http"
	"strings"
	"slices"
	"encoding/json"
	"time"
	"sort"
	"github.com/google/uuid"
	"github.com/tmcausla/chirpy/internal/auth"
	"github.com/tmcausla/chirpy/internal/database"
)

type Chirp struct {
	ID 		  uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
    Body   	  string 	`json:"body"`
	UserID 	  uuid.UUID `json:"user_id"`
}

func (cfg * apiConfig) getChirpByID(w http.ResponseWriter, req *http.Request) {
	chirpID := req.PathValue("chirpID")
	id, err := uuid.Parse(chirpID)
	if err != nil {
		log.Printf("404 Not Found")
		respondWithError(w, 404, "404 Not Found")
		return
	}

	dbChirp, err := cfg.dbQueries.GetChirpByID(req.Context(), id)
	if err != nil {
		log.Printf("404 Not Found, %v\n", err)
		respondWithError(w, 404, "404 Not Found")
		return
	}

	respondWithJSON(w, 200, Chirp{
		ID: dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body: dbChirp.Body,
		UserID: dbChirp.UserID,
	})
}

func (cfg *apiConfig) getChirpsByAuthor(authorID uuid.UUID, req *http.Request) ([]Chirp, error) {
	dbChirps, err := cfg.dbQueries.GetChirpsByAuthor(req.Context(), authorID)
	if err != nil {
		log.Printf("error fetching chirps: %v\n", err)
		return []Chirp{}, err
	}

	chirps := make([]Chirp, len(dbChirps))
	for i, ch := range dbChirps {
		chirps[i] = Chirp{
			ID: ch.ID, 
			CreatedAt: ch.CreatedAt,
			UpdatedAt: ch.UpdatedAt,
			Body: ch.Body,
			UserID: ch.UserID,
		}
	}
	return chirps, nil
}

func sortChirps(chirps []Chirp, sortMethod string) []Chirp {
	if sortMethod == "asc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
		})
		return chirps
	} else if sortMethod == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[j].CreatedAt.Before(chirps[i].CreatedAt)
		})
		return chirps
	}
	log.Printf("unknown sort method: %s\n", sortMethod)
	return chirps
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, req *http.Request) {
	sortMethod := "asc"
	sortQuery := req.URL.Query().Get("sort")
	if sortQuery != "" { sortMethod = sortQuery }

	authorID := req.URL.Query().Get("author_id")
	if authorID != "" {
		id, err := uuid.Parse(authorID)
		if err != nil {
			log.Printf("error parsing author id: %v\n", err)
			respondWithError(w, 404, "404 Not Found")
			return
		}

		authorChirps, err := cfg.getChirpsByAuthor(id, req)
		if err != nil {
			log.Printf("error fetching chirps: %v\n", err)
			respondWithError(w, 404, "404 Not Found")
			return
		}
		authorChirps = sortChirps(authorChirps, sortMethod)

		respondWithJSON(w, 200, authorChirps)
		return
	}

	dbChirps, err := cfg.dbQueries.GetChirps(req.Context())
	if err != nil {
		log.Printf("error fetching chirps: %v\n", err)
		respondWithError(w, 500, "error fetching chirps")
		return
	}

	chirps := make([]Chirp, len(dbChirps))
	for i, ch := range dbChirps {
		chirps[i] = Chirp{
			ID: ch.ID, 
			CreatedAt: ch.CreatedAt,
			UpdatedAt: ch.UpdatedAt,
			Body: ch.Body,
			UserID: ch.UserID,
		}
	}
	chirps = sortChirps(chirps, sortMethod)

	respondWithJSON(w, 200, chirps)
}

func (cfg *apiConfig) postChirp(w http.ResponseWriter, req *http.Request) {
	newChirp := validateChirp(w, req)
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

	dbChirp, err := cfg.dbQueries.CreateChirp(req.Context(), database.CreateChirpParams{
		Body: newChirp.Body,
		UserID: authID,
	})
	if err != nil {
		log.Printf("error creating chirp: %v\n", err)
		respondWithError(w, 500, "error creating chirp")
		return
	}

	respondWithJSON(w, 201, Chirp{
		ID: dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body: dbChirp.Body,
		UserID: dbChirp.UserID,
	})
}

func validateChirp(w http.ResponseWriter, req *http.Request) Chirp {
	type newChirp struct {
		Body   string 	 `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

    decoder := json.NewDecoder(req.Body)
    chirp := newChirp{}
    err := decoder.Decode(&chirp)
    if err != nil {
        log.Printf("error decoding request: %v\n", err)
        respondWithError(w, 400, "error decoding request")
        return Chirp{}
    }

    if len(chirp.Body) > 140 {
        respondWithError(w, 400, "Chirp is too long")
        return Chirp{}
    }

    return Chirp{
		Body: cleanChirp(chirp.Body),
		UserID: chirp.UserID,
	}
}

func cleanChirp(msg string) string {
    badWords := []string{"kerfuffle", "sharbert", "fornax"}
    arr := strings.Split(msg, " ")

    for i := range arr {
        if slices.Contains(badWords, strings.ToLower(arr[i])) {arr[i] = "****"}
    }

    return strings.Join(arr, " ")
}

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, req *http.Request) {
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

	chirpID := req.PathValue("chirpID")
	id, err := uuid.Parse(chirpID)
	if err != nil {
		log.Printf("404 Not Found")
		respondWithError(w, 404, "404 Not Found")
		return
	}

	dbChirp, err := cfg.dbQueries.GetChirpByID(req.Context(), id)
	if err != nil {
		log.Printf("404 Not Found, %v\n", err)
		respondWithError(w, 404, "404 Not Found")
		return
	}

	if dbChirp.UserID != authID {
		log.Printf("user is not author of chirp: %v\n", err)
		respondWithError(w, 403, "403 Unauthorized")
		return
	}

	err = cfg.dbQueries.DeleteChirp(req.Context(), dbChirp.ID)
	if err != nil {
		log.Printf("error deleting chirp: %v\n", err)
		respondWithError(w, 500, "Internal Server Error")
		return
	}

	w.WriteHeader(204)
}