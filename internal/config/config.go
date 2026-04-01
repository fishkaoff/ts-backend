package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	RESTConfig  RESTConfig
	MongoConfig MongoConfig
	JWTConfig   JWTConfig
}

type RESTConfig struct {
	Addr  string `env:"REST_ADDR" env-default:":8080"`
	Rate  int    `env:"LIMITER_RATE" env-default:"200"`
	Burst int    `env:"LIMITER_BURST" env-default:"100"`
}

type MongoConfig struct {
	ClusterUrl string `env:"MONGO_CLUSTER_URL,required"`
}

type JWTConfig struct {
	AccessTokenLifeTime  int    `env:"ACCESS_TOKEN_LIFE_TIME"`  //hours
	RefreshTokenLifeTime int    `env:"REFRESH_TOKEN_LIFE_TIME"` //hours
	SecretKey            string `env:"SECRET_KEY"`
}

func MustLoad() *Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("failed to load config from environment: %v", err)
	}

	return &cfg
}
