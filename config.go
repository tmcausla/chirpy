package main
import (
    "log"
	"fmt"
	"net/http"
	"sync/atomic"
	"github.com/tmcausla/chirpy/internal/database"
)

type apiConfig struct {
    fileserverHits atomic.Int32
    dbQueries      *database.Queries
    platform       string
    tokenSecret    string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        cfg.fileserverHits.Add(1)
        next.ServeHTTP(w, req)
    }) 
}

func (cfg *apiConfig) printHits(w http.ResponseWriter, req *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    w.WriteHeader(http.StatusOK)
    output := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>",
        cfg.fileserverHits.Load())
    w.Write([]byte(output))
}

func (cfg *apiConfig) resetHits(w http.ResponseWriter, req *http.Request) {
    if cfg.platform != "dev" {
        respondWithError(w, 403, "403 Forbidden")
        return
    }
    err := cfg.dbQueries.DeleteUsers(req.Context())
    if err != nil {
        log.Printf("error deleting users: %v", err)
        respondWithError(w, 500, "error deleting users")
        return
    }

    cfg.fileserverHits.Store(0)
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
}
