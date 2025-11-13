package config

import (
    "log"
    "strings"

    "github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	Solana   SolanaConfig
	AWS      AWSConfig
}

type ServerConfig struct {
	Port string
}

type PostgresConfig struct {
	URL string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type SolanaConfig struct {
	RPCEndpoint string
	Network     string // devnet, testnet, mainnet-beta
}

type AWSConfig struct {
	Region  string
	KMSKeyID string
}

func Load() *Config {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./configs")
    viper.AddConfigPath(".")
    viper.AddConfigPath("./internal/config")

    // Environment variables
    viper.AutomaticEnv()
    viper.SetEnvPrefix("APP")
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))

	// Defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("solana.network", "devnet")
	viper.SetDefault("solana.rpc_endpoint", "https://api.devnet.solana.com")
	
    viper.BindEnv("postgres.url", "POSTGRES_URL")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatalf("Error reading config: %v", err)
		}
		// Config file not found; rely on env vars
	}

	return &Config{
		Server: ServerConfig{
			Port: viper.GetString("server.port"),
		},
		Postgres: PostgresConfig{
			URL: viper.GetString("postgres.url"),
		},
		Redis: RedisConfig{
			Addr:     viper.GetString("redis.addr"),
			Password: viper.GetString("redis.password"),
			DB:       viper.GetInt("redis.db"),
		},
		Solana: SolanaConfig{
			RPCEndpoint: viper.GetString("solana.rpc_endpoint"),
			Network:     viper.GetString("solana.network"),
		},
		AWS: AWSConfig{
			Region:   viper.GetString("aws.region"),
			KMSKeyID: viper.GetString("aws.kms_key_id"),
		},
	}
}
