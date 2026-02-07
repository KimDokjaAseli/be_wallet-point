package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerHost     string
	ServerPort     string
	ServerAddress  string
	GinMode        string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	JWTSecret      string
	JWTExpiryHours int
	AllowedOrigins string
	MaxUploadSize  int64
	UploadPath     string
}

func LoadConfig() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Parse JWT expiry hours
	jwtExpiry, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		jwtExpiry = 24
	}

	// Parse max upload size
	maxUploadSize, err := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE", "10485760"), 10, 64)
	if err != nil {
		maxUploadSize = 10485760
	}

	serverHost := getEnv("SERVER_HOST", "0.0.0.0")
	serverPort := getEnv("PORT", "8080")

	return &Config{
		ServerHost:     serverHost,
		ServerPort:     serverPort,
		ServerAddress:  serverHost + ":" + serverPort,
		GinMode:        getEnv("GIN_MODE", "debug"),
		DBHost:         getEnv("DB_HOST", "maglev.proxy.rlwy.net"),
		DBPort:         getEnv("DB_PORT", "35906"),
		DBUser:         getEnv("DB_USER", "root"),
		DBPassword:     getEnv("DB_PASSWORD", "rgkMnrENbVTZwmfoaFqHVXFoJXqRrRTW"),
		DBName:         getEnv("DB_NAME", "railway"),
		JWTSecret:      getEnv("JWT_SECRET", "H6RoFvCDVvlXU33SXXsi2anbRF/mafbH9O1+QWoulji6n8xtgiVXeorrSJTwr83LDVVX8wxYexICnyCyjpg=="),
		JWTExpiryHours: jwtExpiry,
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
		MaxUploadSize:  maxUploadSize,
		UploadPath:     getEnv("UPLOAD_PATH", "./uploads"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
