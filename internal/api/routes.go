package api

import "net/http"

// NewMux builds the HTTP server mux with all routes registered.
func NewMux(cfg *Config) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/app/", http.StripPrefix("/app/", cfg.middlewareMetrics(http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", getReadyStatus)

	mux.HandleFunc("GET /admin/metrics", cfg.getHitCount)
	mux.HandleFunc("POST /admin/reset", cfg.resetAPI)

	mux.HandleFunc("POST /api/chirps", cfg.createChirps)
	mux.HandleFunc("GET /api/chirps", cfg.getChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.getChirps)

	mux.HandleFunc("POST /api/users", cfg.createUser)
	mux.HandleFunc("POST /api/login", cfg.login)

	mux.HandleFunc("POST /api/refresh", cfg.refreshJWT)
	mux.HandleFunc("POST /api/revoke", cfg.revokeRefreshToken)
	return mux
}
