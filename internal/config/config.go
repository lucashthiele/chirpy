package config

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type ApiConfig struct {
	FileServerHits *atomic.Int32
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
