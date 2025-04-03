package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server struct {
		Port    string `mapstructure:"port"`
		Host    string `mapstructure:"host"`
		DataDir string `mapstructure:"data_dir"`
	} `mapstructure:"server"`
	Session struct {
		Secret        string   `mapstructure:"secret"`
		Name          string   `mapstructure:"name"`
		AllowedEmails []string `mapstructure:"allowed_emails"`
		Secure        bool     `mapstructure:"secure"`
	} `mapstructure:"session"`
	Google struct {
		ClientID      string   `mapstructure:"client_id"`
		ClientSecret  string   `mapstructure:"client_secret"`
		RedirectURL   string   `mapstructure:"redirect_url"`
		AllowedEmails []string `mapstructure:"allowed_emails"`
	} `mapstructure:"google"`
	GitHub struct {
		Token      string `mapstructure:"token"`
		Owner      string `mapstructure:"owner"`
		Repository string `mapstructure:"repository"`
		Branch     string `mapstructure:"branch"`
	} `mapstructure:"github"`
	Redis struct {
		Address           string `mapstructure:"address"`
		Password          string `mapstructure:"password"`
		DB                int    `mapstructure:"db"`
		Enabled           bool   `mapstructure:"enabled"`
		ExpirationSeconds int    `mapstructure:"expiration_seconds"`
	} `mapstructure:"redis"`
	Wiki struct {
		MaxCategoryLevel int `mapstructure:"max_category_level"`
	} `mapstructure:"wiki"`
}

var AppConfig Config

// LoadConfig loads the configuration from the environment file
func LoadConfig() error {
	viper.SetConfigName("env")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Set default Redis values if not specified
	if AppConfig.Redis.Address == "" {
		AppConfig.Redis.Address = "localhost:6379"
	}
	if AppConfig.Redis.DB == 0 {
		AppConfig.Redis.DB = 0
	}
	if !AppConfig.Redis.Enabled {
		AppConfig.Redis.Enabled = true // Enable Redis by default
	}
	if AppConfig.Redis.ExpirationSeconds == 0 {
		AppConfig.Redis.ExpirationSeconds = 900 // 15 minutes default
	}

	// Set default Wiki values if not specified
	if AppConfig.Wiki.MaxCategoryLevel == 0 {
		AppConfig.Wiki.MaxCategoryLevel = 4 // Default to 4 levels
	}

	return nil
}

// GetConfig returns the current configuration
func GetConfig() *Config {
	return &AppConfig
}

// GetServerPort returns the server port
func GetServerPort() string {
	return AppConfig.Server.Port
}

// GetServerHost returns the server host
func GetServerHost() string {
	return AppConfig.Server.Host
}

// GetMaxCategoryLevel returns the maximum allowed category nesting level
func GetMaxCategoryLevel() int {
	return AppConfig.Wiki.MaxCategoryLevel
}
