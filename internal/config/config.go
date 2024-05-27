package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	defaultGRPCPort = "443"

	defaultHTTPPort               = "8000"
	defaultHTTPRWTimeout          = 10 * time.Second
	defaultHTTPMaxHeaderMegabytes = 1
	defaultAccessTokenTTL         = 15 * time.Minute
	defaultActivationTokenTTL     = 4 * time.Hour
	defaultRefreshTokenTTL        = 12 * time.Hour

	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type (
	Config struct {
		Environment string
		Application string
		Mongo       MongoConfig `mapstructure:"mongo"`
		Auth        AuthConfig  `mapstructure:"auth"`
		GRPC        GRPCConfig  `mapstructure:"grpc"`
	}

	MongoConfig struct {
		URI      string
		User     string
		Password string
		Name     string `mapstructure:"databaseName"`
	}

	AuthConfig struct {
		JWT                JWTConfig
		RefreshTokenTTL    time.Duration `mapstructure:"refreshTokenTTL"`
		ActivationTokenTTL time.Duration `mapstructure:"activationTokenTTL"`
		ActivationCodeTTL  time.Duration `mapstructure:"activationCodeTTL"`
		PasswordCost       int
	}

	JWTConfig struct {
		AccessTokenTTL time.Duration `mapstructure:"accessTokenTTL"`
		SigningKey     string
	}

	GRPCConfig struct {
		Host    string        `mapstructure:"host"`
		Port    string        `mapstructure:"port"`
		Timeout time.Duration `mapstructure:"timeout"`
	}
)

func Init(configsDir, envDir string) (*Config, error) {
	var cfg Config
	populateDefaults()
	loadEnvVariables(envDir)
	if err := parseConfigFile(configsDir); err != nil {
		return nil, err
	}

	if err := unmarshal(&cfg); err != nil {
		return nil, err
	}

	setFromEnv(&cfg)
	return &cfg, nil
}

func unmarshal(cfg *Config) error {

	if err := viper.UnmarshalKey("mongo", &cfg.Mongo); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("grpc", &cfg.GRPC); err != nil {
		return err
	}
	if err := viper.UnmarshalKey("auth", &cfg.Auth); err != nil {
		return err
	}
	return viper.UnmarshalKey("auth", &cfg.Auth.JWT)
}

func setFromEnv(cfg *Config) {
	cfg.Mongo.URI = os.Getenv("MONGO_URI")
	cfg.Mongo.User = os.Getenv("MONGO_USER")
	cfg.Mongo.Password = os.Getenv("MONGO_PASS")

	cfg.Auth.PasswordCost, _ = strconv.Atoi(os.Getenv("PASSWORD_SALT"))
	cfg.Auth.JWT.SigningKey = os.Getenv("JWT_SIGNING_KEY")

	cfg.GRPC.Host = os.Getenv("GRPC_HOST")

	cfg.Environment = envDev
	cfg.Application = os.Getenv("APP")
}

func parseConfigFile(folder string) error {
	viper.AddConfigPath(folder)
	viper.SetConfigName("main")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return viper.MergeInConfig()
}

func loadEnvVariables(envPath string) {
	err := godotenv.Load(envPath)

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

}

func populateDefaults() {
	viper.SetDefault("grpc.port", defaultGRPCPort)
	viper.SetDefault("http.port", defaultHTTPPort)
	viper.SetDefault("http.max_header_megabytes", defaultHTTPMaxHeaderMegabytes)
	viper.SetDefault("http.timeouts.read", defaultHTTPRWTimeout)
	viper.SetDefault("http.timeouts.write", defaultHTTPRWTimeout)
	viper.SetDefault("auth.accessTokenTTL", defaultAccessTokenTTL)
	viper.SetDefault("auth.refreshTokenTTL", defaultRefreshTokenTTL)
	viper.SetDefault("auth.activationTokenTTL", defaultActivationTokenTTL)
}
