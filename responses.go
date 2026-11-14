package main
import (
	"log"
	"net/http"
	"encoding/json"
)

func respondReady(w http.ResponseWriter, req *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
    type errorResp struct {
        Error string `json:"error"`
    }
    errResp := errorResp{
        Error: msg,
    }

    data, err := json.Marshal(errResp)
    if err != nil {
        log.Printf("Error marshalling error JSON: %s", err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(data)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    data, err := json.Marshal(payload)
    if err != nil {
        log.Printf("error marshaling json response: %s", err)
        respondWithError(w, 500, "error marshaling json response")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(data)
}