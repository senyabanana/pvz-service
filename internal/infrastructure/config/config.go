package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	ServerPort       string `mapstructure:"SERVER_PORT"`
	GRPCPort         string `mapstructure:"GRPC_PORT"`
	PostgresHost     string `mapstructure:"POSTGRES_HOST"`
	PostgresPort     string `mapstructure:"POSTGRES_PORT"`
	PostgresUser     string `mapstructure:"POSTGRES_USER"`
	PostgresPassword string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresDB       string `mapstructure:"POSTGRES_DB"`
	SSLMode          string `mapstructure:"SSLMODE"`
	JWTSecretKey     string `mapstructure:"JWTKEY"`
}

func LoadConfig(path string) (cfg *Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&cfg)
	return
}
