package database

// Config struct
type Config struct {
	Server   string `yaml:"server" json:"server"`
	Port     int    `yaml:"port" json:"port"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	Name     string `yaml:"name" json:"name"`
}

// PostgresConfig contains fields for PostgreSQL configuration
type PostgresConfig struct {
	Config `yaml:"config" json:"config"`
	UseSSL bool `yaml:"use_ssl" json:"use_ssl"`
}
