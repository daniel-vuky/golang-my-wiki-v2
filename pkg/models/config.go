package models

// Config represents the application configuration
type Config struct {
	Google struct {
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
		RedirectURL  string `yaml:"redirect_url"`
	} `yaml:"google"`

	Session struct {
		Secret        string   `yaml:"secret"`
		Name          string   `yaml:"name"`
		Secure        bool     `yaml:"secure"`
		AllowedEmails []string `yaml:"allowed_emails"`
	} `yaml:"session"`

	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`

	GitHub struct {
		Token      string `yaml:"token"`
		Owner      string `yaml:"owner"`
		Repository string `yaml:"repository"`
		Branch     string `yaml:"branch"`
	} `yaml:"github"`

	StorageMode string `yaml:"storage_mode"`
}
