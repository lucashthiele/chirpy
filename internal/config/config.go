package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/lucashthiele/chirpy/internal/auth"
	"github.com/lucashthiele/chirpy/internal/database"
	"github.com/lucashthiele/chirpy/pkg/response"
)

type contextKey string

const UserIDKey contextKey = "userID"

type ApiConfig struct {
	Platform       string
	FileServerHits *atomic.Int32
	Db             *database.Queries
	AppSecret      string
	PolkaKey       string
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
			Platform:       os.Getenv("PLATFORM"),
			FileServerHits: &atomic.Int32{},
			Db:             db,
			AppSecret:      os.Getenv("APP_SECRET"),
			PolkaKey:       os.Getenv("POLKA_KEY"),
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

func (cfg *ApiConfig) MiddlewareAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		token, err := auth.GetBearerToken(&req.Header)
		if err != nil {
			response.RespondWithError(resp, http.StatusUnauthorized, err.Error())
			return
		}

		userId, err := auth.ValidateJWT(token, cfg.AppSecret)
		if err != nil {
			response.RespondWithError(resp, http.StatusUnauthorized, "Unauthorized")
			return
		}
		ctx := context.WithValue(req.Context(), UserIDKey, userId)

		next.ServeHTTP(resp, req.WithContext(ctx))
	})
}

func (cfg *ApiConfig) MiddlewarePolka(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		APIKey, err := auth.GetAPIKey(&req.Header)
		if err != nil {
			response.RespondWithError(resp, http.StatusUnauthorized, err.Error())
			return
		}

		if cfg.PolkaKey != APIKey {
			response.RespondWithError(resp, http.StatusUnauthorized, "Unauthorized")
			return
		}

		next.ServeHTTP(resp, req)
	})
}

func (cfg *ApiConfig) HandleReset() http.HandlerFunc {
	return http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
		if cfg.Platform != "dev" {
			response.RespondWithError(resp, http.StatusForbidden, "Forbidden")
			return
		}
		err := cfg.Db.DeleteAllUsers(r.Context())
		if err != nil {
			response.RespondWithInternalServerError(resp, err)
			return
		}
		cfg.FileServerHits.Store(0)
		log.Println("Reset endpoint called. Everything was wiped")
		resp.WriteHeader(http.StatusOK)
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
