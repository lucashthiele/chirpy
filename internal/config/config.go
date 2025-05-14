package config

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/lucashthiele/chirpy/internal/database"
)

type ApiConfig struct {
	FileServerHits *atomic.Int32
	Db             *database.Queries
}

var instance *ApiConfig

func createDatabaseInstance() (*database.Queries, error) {
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return &database.Queries{}, fmt.Errorf("error opening database connection: %s", err.Error())
	}

	return database.New(db), nil
}

func New() (*ApiConfig, error) {
	if instance == nil {
		db, err := createDatabaseInstance()
		if err != nil {
			return &ApiConfig{}, nil
		}

		instance = &ApiConfig{
			FileServerHits: &atomic.Int32{},
			Db:             db,
		}
		instance.FileServerHits.Store(0)
	}

	return instance, nil
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) HandleReset() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileServerHits.Store(0)
		w.WriteHeader(http.StatusOK)
	})
}

func (cfg *ApiConfig) HandleMetrics() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricsResp := fmt.Sprintf(`<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>`, cfg.FileServerHits.Load())
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(metricsResp))
	})
}
