package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/lucashthiele/chirpy/internal/config"
	"github.com/lucashthiele/chirpy/internal/handlers"
)

const port string = "42069"

func getFilepathRoot() http.Dir {
	return http.Dir(".")
}

func configureRoutes(mux *http.ServeMux, cfg *config.ApiConfig) {
	handler := http.StripPrefix("/app/", http.FileServer(getFilepathRoot()))
	mux.Handle("/app/", cfg.MiddlewareMetricsInc(handler))

	mux.HandleFunc("GET /admin/metrics", cfg.HandleMetrics())
	mux.HandleFunc("POST /admin/reset", cfg.HandleReset())

	mux.HandleFunc("GET /api/healthz", handlers.HandleHealthz)
	mux.HandleFunc("POST /api/validate_chirp", handlers.HandleValidateChirp)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading env: %s", err.Error())
		return
	}

	cfg, err := config.New()
	if err != nil {
		fmt.Printf("Error setting configuration: %s", err.Error())
		return
	}
	mux := http.NewServeMux()

	configureRoutes(mux, cfg)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", getFilepathRoot(), port)
	err = server.ListenAndServe()
	if err != nil {
		fmt.Printf("Error starting server: %s", err.Error())
		return
	}
}
