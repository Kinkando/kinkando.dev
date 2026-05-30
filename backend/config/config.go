package config

import (
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Port                    string `env:"PORT" envDefault:"8080"`
	AppEnv                  string `env:"APP_ENV" envDefault:"development"`
	PostgresDSN             string `env:"POSTGRES_DSN,required"`
	MongoURI                string `env:"MONGO_URI,required"`
	MongoDB                 string `env:"MONGO_DB,required"`
	FirebaseCredentials string `env:"FIREBASE_CREDENTIALS,required"`
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
