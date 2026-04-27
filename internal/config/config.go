package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv             string
	HTTPPort           int
	DatabaseURL        string
	JWTSecret          string
	TokenTTL           time.Duration
	PublicBaseURL      string
	UploadDir          string
	MigrationsDir      string
	AutoMigrate        bool
	MaxUploadSizeBytes int64
}

func Load() (Config, error) {
	_ = godotenv.Load()

	portValue := getenv("HTTP_PORT", "3000")
	port, err := strconv.Atoi(portValue)
	if err != nil {
		return Config{}, fmt.Errorf("parse HTTP_PORT: %w", err)
	}

	tokenTTLMinutes, err := strconv.Atoi(getenv("TOKEN_TTL_MINUTES", "120"))
	if err != nil {
		return Config{}, fmt.Errorf("parse TOKEN_TTL_MINUTES: %w", err)
	}

	autoMigrate, err := strconv.ParseBool(getenv("AUTO_MIGRATE", "true"))
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTO_MIGRATE: %w", err)
	}

	maxUploadSizeBytes, err := strconv.ParseInt(getenv("MAX_UPLOAD_SIZE_BYTES", "5242880"), 10, 64)
	if err != nil {
		return Config{}, fmt.Errorf("parse MAX_UPLOAD_SIZE_BYTES: %w", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	return Config{
		AppEnv:             getenv("APP_ENV", "development"),
		HTTPPort:           port,
		DatabaseURL:        databaseURL,
		JWTSecret:          jwtSecret,
		TokenTTL:           time.Duration(tokenTTLMinutes) * time.Minute,
		PublicBaseURL:      getenv("PUBLIC_BASE_URL", fmt.Sprintf("http://localhost:%d", port)),
		UploadDir:          getenv("UPLOAD_DIR", "storage"),
		MigrationsDir:      getenv("MIGRATIONS_DIR", "db/migrations"),
		AutoMigrate:        autoMigrate,
		MaxUploadSizeBytes: maxUploadSizeBytes,
	}, nil
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
