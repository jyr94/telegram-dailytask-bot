// config/config.go
package config

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken     string
	GoogleSheetID     string
	CredentialsJSON   string
	FirebaseProjectId string
	CredentialsBase64 string
}

func Load() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ No .env file found, using system environment variables.")
	}
	cfg := Config{
		TelegramToken:     os.Getenv("TELEGRAM_TOKEN"),
		CredentialsJSON:   os.Getenv("GOOGLE_CREDENTIALS_FILE"),
		FirebaseProjectId: os.Getenv("FIREBASE_PROJECT_ID"),
		CredentialsBase64: os.Getenv("GOOGLE_CREDENTIALS_BASE64"),
	}
	filepath := WriteCredentialsToTempFile(cfg.CredentialsBase64)
	fmt.Println(filepath)
	cfg.CredentialsJSON = filepath

	if cfg.TelegramToken == "" || cfg.CredentialsJSON == "" {
		log.Fatal("❌ Missing required environment variables")
	}
	return cfg
}

func WriteCredentialsToTempFile(b64 string) string {

	data, _ := base64.StdEncoding.DecodeString(b64)

	tmpFile, err := os.CreateTemp("", "firebase-*.json")
	if err != nil {
		log.Fatalf("❌ Failed to create temp file: %v", err)
	}
	tmpFile.Write(data)
	tmpFile.Close()

	return tmpFile.Name()
}
