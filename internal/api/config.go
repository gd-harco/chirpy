package api

import (
	"sync/atomic"

	"github.com/gd-harco/chirpy/internal/database"
)

// Config holds the shared state used by API handlers.
type Config struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secretKey      string
}

// NewConfig builds a Config ready to be used to construct the server mux.
func NewConfig(db *database.Queries, platform string, secretKey string) *Config {
	return &Config{
		db:        db,
		platform:  platform,
		secretKey: secretKey,
	}
}
