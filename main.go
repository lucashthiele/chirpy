package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/lucashthiele/chirpy/internal/config"
	"github.com/lucashthiele/chirpy/internal/database"
	"github.com/lucashthiele/chirpy/internal/handlers"
)

const port string = "42069"

func getFilepathRoot() http.Dir {
	return http.Dir(".")
}

func createDatabaseInstance() error {
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("error opening database connection: %s", err.Error())
	}

	_ = database.New(db)
	return nil
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

	err = createDatabaseInstance()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	cfg := &config.ApiConfig{
		FileServerHits: &atomic.Int32{},
	}
	cfg.FileServerHits.Store(0)
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
