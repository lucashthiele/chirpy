package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/lucashthiele/chirpy/internal/database"
)

const port string = "42069"

type apiConfig struct {
	fileServerHits *atomic.Int32
}

type parameters struct {
	Body string `json:"body"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type bodyResp struct {
	Body string `json:"cleaned_body"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetric() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricsResp := fmt.Sprintf(`<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>`, cfg.fileServerHits.Load())
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(metricsResp))
	})
}

func (cfg *apiConfig) handlerReset() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Store(0)
		w.WriteHeader(http.StatusOK)
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.WriteHeader(code)
	resp, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error encoding response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)

	errResp := errorResponse{
		Error: msg,
	}
	resp, err := json.Marshal(errResp)
	if err != nil {
		log.Printf("Error encoding response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func handleHealthz(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("OK"))
}

func validateChirp(str string) error {
	if len(str) > 140 {
		return fmt.Errorf("chirp is too long")
	}

	return nil
}

func removeProfaneWords(str string) string {
	badWords := map[string]string{
		"kerfuffle": "****",
		"sharbert":  "****",
		"fornax":    "****",
	}

	splittedWords := strings.Split(str, " ")
	for i, word := range splittedWords {
		if replace, found := badWords[strings.ToLower(word)]; found {
			splittedWords[i] = replace
		}
	}

	return strings.Join(splittedWords, " ")
}

func handleValidateChirp(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding body: %s", err)
		respondWithError(res, http.StatusInternalServerError, "something went wrong")
		return
	}

	err = validateChirp(params.Body)
	if err != nil {
		respondWithError(res, http.StatusBadRequest, err.Error())
		return
	}

	cleanedBody := removeProfaneWords(params.Body)

	bodyResp := bodyResp{
		Body: cleanedBody,
	}

	respondWithJSON(res, http.StatusOK, bodyResp)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading env: %s", err.Error())
		return
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Error opening database connection: %s", err.Error())
		return
	}

	_ = database.New(db)

	cfg := &apiConfig{
		fileServerHits: &atomic.Int32{},
	}
	cfg.fileServerHits.Store(0)
	mux := http.NewServeMux()
	filepathRoot := http.Dir(".")

	handler := http.StripPrefix("/app/", http.FileServer(filepathRoot))
	mux.Handle("/app/", cfg.middlewareMetricsInc(handler))

	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetric())
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset())

	mux.HandleFunc("GET /api/healthz", handleHealthz)
	mux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	server.ListenAndServe()
}
