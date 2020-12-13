package database

// Config struct
type Config struct {
	Server   string `yaml:"server" json:"server"`
	Port     int    `yaml:"port" json:"port"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	Name     string `yaml:"name" json:"name"`
}
