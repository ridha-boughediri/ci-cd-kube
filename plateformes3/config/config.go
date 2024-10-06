package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config struct pour stocker les paramètres de configuration
type Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	StoragePath     string
}

// LoadConfig charge la configuration à partir du fichier .env
func LoadConfig() Config {
	// Chargement des variables d'environnement depuis le fichier .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using default values")
	}

	// Récupération des valeurs des variables d'environnement
	cfg := Config{
		AccessKeyID:     os.Getenv("ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("SECRET_ACCESS_KEY"),
		Region:          os.Getenv("REGION"),
		StoragePath:     os.Getenv("STORAGE_PATH"),
	}

	return cfg
}
