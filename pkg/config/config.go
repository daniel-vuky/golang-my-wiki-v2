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
		AllowedEmails []string `mapstructure:"allowed_emails"`
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
	StorageMode string `mapstructure:"storage_mode"`
}

var AppConfig Config

// LoadConfig loads the configuration from the environment file
func LoadConfig() error {
	viper.SetConfigName("env")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Set default storage mode if not specified
	if AppConfig.StorageMode == "" {
		AppConfig.StorageMode = "local"
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
