package configs

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPaths []string, configName string, configType string) (Config, error) {
	v := viper.New()
	v.SetConfigName(configName)
	v.SetConfigType(configType)

	for _, path := range configPaths {
		v.AddConfigPath(path)
	}

	// Set defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("rpc.port", 8888)
	v.SetDefault("database.type", "sqlite")
	v.SetDefault("database.path", "git_sync.db")
	v.SetDefault("webhook.secret", "my-secret-key")
	v.SetDefault("webhook.rate_limit", 100)
	v.SetDefault("webhook.ip_whitelist", []string{})

	// Storage defaults (本地存储优先)
	v.SetDefault("storage.type", "local")
	v.SetDefault("storage.local_path", "./storage")
	v.SetDefault("storage.repo_bucket", "repos")
	v.SetDefault("storage.ssh_key_bucket", "ssh-keys")
	v.SetDefault("storage.audit_log_bucket", "audit-logs")
	v.SetDefault("storage.backup_bucket", "backups")
	v.SetDefault("storage.use_ssl", false)

	// Lock defaults (内存锁优先)
	v.SetDefault("lock.type", "memory")
	v.SetDefault("lock.redis_db", 0)

	// Environment variables override
	// 支持环境变量覆盖，如 STORAGE_TYPE, LOCK_REDIS_ADDR 等
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found, using defaults")
		} else {
			return Config{}, err
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}
