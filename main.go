package main
import (
    "log"
    "net/http"
    "time"
    "os"
    "database/sql"
    "github.com/joho/godotenv"
    "github.com/tmcausla/chirpy/internal/database"

    _ "github.com/lib/pq"
)

func main() {
    godotenv.Load()
    dbURL := os.Getenv("DB_URL")
    plat := os.Getenv("PLATFORM")
    tokSec := os.Getenv("TOKEN_SECRET")

    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Printf("error connecting to database: %v", err)
        os.Exit(1)
    }

    apiCfg := apiConfig{
        dbQueries: database.New(db),
        platform: plat,
        tokenSecret: tokSec,
    }

    mux := http.NewServeMux()
    mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
    mux.HandleFunc("GET /api/healthz", respondReady)
    mux.HandleFunc("GET /admin/metrics", apiCfg.printHits)
    mux.HandleFunc("POST /admin/reset", apiCfg.resetHits)
    mux.HandleFunc("POST /api/users", apiCfg.createUser)
    mux.HandleFunc("POST /api/chirps", apiCfg.postChirp)
    mux.HandleFunc("GET /api/chirps", apiCfg.getChirps)
    mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirpByID)
    mux.HandleFunc("POST /api/login", apiCfg.login)
    mux.HandleFunc("POST /api/refresh", apiCfg.refresh)
    mux.HandleFunc("POST /api/revoke", apiCfg.revoke)
    mux.HandleFunc("PUT /api/users", apiCfg.updateUser)
    mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.deleteChirp)

    s := &http.Server{
        Addr: ":8080",
        Handler: mux,
        ReadTimeout: 10 * time.Second,
        WriteTimeout: 10 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }

    log.Fatal(s.ListenAndServe())
}
