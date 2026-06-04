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

	// Gemini API — powers natural-language processing in the LINE webhook.
	GeminiAPIKey  string `env:"GEMINI_API_KEY,required"`
	GeminiModel   string `env:"GEMINI_MODEL" envDefault:"gemini-2.0-flash"`
	GeminiTTSModel string `env:"GEMINI_TTS_MODEL" envDefault:"gemini-2.5-flash-preview-tts"`

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
