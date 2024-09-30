// config/config.go
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config structure pour stocker les configurations
type Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	StoragePath     string
}

// LoadConfig charge les variables d'environnement depuis le fichier .env
func LoadConfig() Config {
	// Charger le fichier .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Aucun fichier .env trouvé, utilisation des valeurs par défaut")
	}

	cfg := Config{
		AccessKeyID:     os.Getenv("ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("SECRET_ACCESS_KEY"),
		Region:          os.Getenv("REGION"),
		StoragePath:     os.Getenv("STORAGE_PATH"),
	}

	// Définir des valeurs par défaut si nécessaire
	if cfg.AccessKeyID == "" {
		cfg.AccessKeyID = "default_access_key"
	}
	if cfg.SecretAccessKey == "" {
		cfg.SecretAccessKey = "default_secret_key"
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	if cfg.StoragePath == "" {
		cfg.StoragePath = "./data/"
	}

	return cfg
}
