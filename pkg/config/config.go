package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Google struct {
		ClientID     string `mapstructure:"client_id"`
		ClientSecret string `mapstructure:"client_secret"`
		RedirectURL  string `mapstructure:"redirect_url"`
	} `mapstructure:"google"`

	Session struct {
		Secret        string   `mapstructure:"secret"`
		Name          string   `mapstructure:"name"`
		Secure        bool     `mapstructure:"secure"`
		AllowedEmails []string `mapstructure:"allowed_emails"`
	} `mapstructure:"session"`

	Server struct {
		Port string `mapstructure:"port"`
		Host string `mapstructure:"host"`
	} `mapstructure:"server"`
}

var AppConfig Config

func LoadConfig() error {
	viper.SetConfigName("env")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Read environment variables
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return fmt.Errorf("config file not found: %w", err)
		}
		return fmt.Errorf("error reading config file: %w", err)
	}

	// Unmarshal config into struct
	if err := viper.Unmarshal(&AppConfig); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate required fields
	if AppConfig.Google.ClientID == "" {
		return fmt.Errorf("google client ID is required")
	}
	if AppConfig.Google.ClientSecret == "" {
		return fmt.Errorf("google client secret is required")
	}
	if AppConfig.Session.Secret == "" {
		return fmt.Errorf("session secret is required")
	}

	return nil
}

// GetGoogleClientID returns the Google OAuth client ID
func GetGoogleClientID() string {
	return AppConfig.Google.ClientID
}

// GetGoogleClientSecret returns the Google OAuth client secret
func GetGoogleClientSecret() string {
	return AppConfig.Google.ClientSecret
}

// GetGoogleRedirectURL returns the Google OAuth redirect URL
func GetGoogleRedirectURL() string {
	return AppConfig.Google.RedirectURL
}

// GetSessionSecret returns the session secret
func GetSessionSecret() string {
	return AppConfig.Session.Secret
}

// GetSessionName returns the session name
func GetSessionName() string {
	return AppConfig.Session.Name
}

// GetSessionSecure returns whether the session cookie should be secure
func GetSessionSecure() bool {
	return AppConfig.Session.Secure
}

// GetServerPort returns the server port
func GetServerPort() string {
	return AppConfig.Server.Port
}

// GetServerHost returns the server host
func GetServerHost() string {
	return AppConfig.Server.Host
}

// GetAllowedEmails returns the list of allowed email addresses
func GetAllowedEmails() []string {
	return AppConfig.Session.AllowedEmails
}
