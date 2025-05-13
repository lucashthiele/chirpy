package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
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

type validResponse struct {
	Valid bool `json:"valid"`
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

func handleHealthz(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("OK"))
}

func handleValidateChirp(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		errResp := errorResponse{
			Error: "Something went wrong",
		}
		resp, err := json.Marshal(errResp)
		if err != nil {
			log.Printf("Error enconding response: %s", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("Error decoding parameters: %s", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write(resp)
		return
	}

	if len(params.Body) > 140 {
		res.WriteHeader(http.StatusBadRequest)

		errResp := errorResponse{
			Error: "Chirp is too long",
		}
		resp, err := json.Marshal(errResp)
		if err != nil {
			log.Printf("Error enconding response: %s", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.Write(resp)
		return
	}

	res.WriteHeader(http.StatusOK)
	validResp := validResponse{
		Valid: true,
	}
	resp, err := json.Marshal(validResp)
	if err != nil {
		log.Printf("Error enconding response: %s", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Write(resp)
}

func main() {
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
