package config

import (
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Port                string `env:"PORT" envDefault:"8080"`
	AppEnv              string `env:"APP_ENV" envDefault:"development"`
	PostgresDSN         string `env:"POSTGRES_DSN,required"`
	MongoURI            string `env:"MONGO_URI,required"`
	MongoDB             string `env:"MONGO_DB,required"`
	FirebaseCredentials string `env:"FIREBASE_CREDENTIALS,required"`
	// MCPUserFirebaseUID and MCPAuthToken enable the /mcp endpoint on the HTTP
	// server. Both must be set; if either is empty, MCP is disabled.
	// MCPUserFirebaseUID is also used by the LINE webhook to identify the
	// single app user to write records for.
	MCPUserFirebaseUID string `env:"MCP_USER_FIREBASE_UID,required"`
	MCPAuthToken       string `env:"MCP_AUTH_TOKEN"`

	// Gemini API — powers natural-language processing in the LINE webhook.
	GeminiAPIKey string `env:"GEMINI_API_KEY,required"`
	GeminiModel  string `env:"GEMINI_MODEL" envDefault:"gemini-2.0-flash"`

	// LINE Messaging API credentials.
	LineChannelID          string `env:"LINE_CHANNEL_ID,required"`
	LineChannelSecret      string `env:"LINE_CHANNEL_SECRET,required"`
	LineChannelAccessToken string `env:"LINE_CHANNEL_ACCESS_TOKEN,required"`
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	return &cfg
}
