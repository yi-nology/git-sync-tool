package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Webhook  WebhookConfig  `mapstructure:"webhook"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type DatabaseConfig struct {
	Type     string `mapstructure:"type"` // sqlite, mysql, postgres
	DSN      string `mapstructure:"dsn"`  // Data Source Name (for mysql/postgres)
	Path     string `mapstructure:"path"` // For SQLite
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type WebhookConfig struct {
	Secret      string   `mapstructure:"secret"`
	RateLimit   int      `mapstructure:"rate_limit"`
	IPWhitelist []string `mapstructure:"ip_whitelist"`
}

var (
	GlobalConfig Config

	// Keep backward compatibility
	WebhookSecret      = "my-secret-key"
	WebhookRateLimit   = 100
	WebhookIPWhitelist = []string{}
	DebugMode          = false
)

func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.path", "git_sync.db")
	viper.SetDefault("webhook.secret", "my-secret-key")
	viper.SetDefault("webhook.rate_limit", 100)
	viper.SetDefault("webhook.ip_whitelist", []string{})

	// Environment variables override
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found, using defaults")
		} else {
			log.Fatalf("Fatal error config file: %s \n", err)
		}
	}

	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	// Update global variables for backward compatibility
	WebhookSecret = GlobalConfig.Webhook.Secret
	WebhookRateLimit = GlobalConfig.Webhook.RateLimit
	WebhookIPWhitelist = GlobalConfig.Webhook.IPWhitelist

	// Manual override for old ENV vars
	if secret := os.Getenv("WEBHOOK_SECRET"); secret != "" {
		WebhookSecret = secret
		GlobalConfig.Webhook.Secret = secret
	}

	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		GlobalConfig.Database.Path = dbPath
	}
}
