package database

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfigJSONMarshaling(t *testing.T) {
	config := Config{
		Server:   "localhost",
		Port:     5432,
		Username: "testuser",
		Password: "testpass",
		Name:     "testdb",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal Config to JSON: %v", err)
	}

	// Unmarshal back
	var decoded Config
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Config from JSON: %v", err)
	}

	// Verify fields
	if decoded.Server != config.Server {
		t.Errorf("Server = %q, want %q", decoded.Server, config.Server)
	}
	if decoded.Port != config.Port {
		t.Errorf("Port = %d, want %d", decoded.Port, config.Port)
	}
	if decoded.Username != config.Username {
		t.Errorf("Username = %q, want %q", decoded.Username, config.Username)
	}
	if decoded.Password != config.Password {
		t.Errorf("Password = %q, want %q", decoded.Password, config.Password)
	}
	if decoded.Name != config.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, config.Name)
	}
}

func TestConfigYAMLMarshaling(t *testing.T) {
	config := Config{
		Server:   "localhost",
		Port:     5432,
		Username: "testuser",
		Password: "testpass",
		Name:     "testdb",
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal Config to YAML: %v", err)
	}

	// Unmarshal back
	var decoded Config
	if err := yaml.Unmarshal(yamlData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Config from YAML: %v", err)
	}

	// Verify fields
	if decoded.Server != config.Server {
		t.Errorf("Server = %q, want %q", decoded.Server, config.Server)
	}
	if decoded.Port != config.Port {
		t.Errorf("Port = %d, want %d", decoded.Port, config.Port)
	}
	if decoded.Username != config.Username {
		t.Errorf("Username = %q, want %q", decoded.Username, config.Username)
	}
	if decoded.Password != config.Password {
		t.Errorf("Password = %q, want %q", decoded.Password, config.Password)
	}
	if decoded.Name != config.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, config.Name)
	}
}

func TestConfigJSONFieldNames(t *testing.T) {
	jsonStr := `{"server":"myserver","port":1433,"username":"admin","password":"secret","name":"mydb"}`

	var config Config
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if config.Server != "myserver" {
		t.Errorf("Server = %q, want %q", config.Server, "myserver")
	}
	if config.Port != 1433 {
		t.Errorf("Port = %d, want %d", config.Port, 1433)
	}
	if config.Username != "admin" {
		t.Errorf("Username = %q, want %q", config.Username, "admin")
	}
	if config.Password != "secret" {
		t.Errorf("Password = %q, want %q", config.Password, "secret")
	}
	if config.Name != "mydb" {
		t.Errorf("Name = %q, want %q", config.Name, "mydb")
	}
}

func TestConfigYAMLFieldNames(t *testing.T) {
	yamlStr := `server: myserver
port: 1433
username: admin
password: secret
name: mydb`

	var config Config
	if err := yaml.Unmarshal([]byte(yamlStr), &config); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	if config.Server != "myserver" {
		t.Errorf("Server = %q, want %q", config.Server, "myserver")
	}
	if config.Port != 1433 {
		t.Errorf("Port = %d, want %d", config.Port, 1433)
	}
	if config.Username != "admin" {
		t.Errorf("Username = %q, want %q", config.Username, "admin")
	}
	if config.Password != "secret" {
		t.Errorf("Password = %q, want %q", config.Password, "secret")
	}
	if config.Name != "mydb" {
		t.Errorf("Name = %q, want %q", config.Name, "mydb")
	}
}

func TestConfigEmptyValues(t *testing.T) {
	config := Config{}

	if config.Server != "" {
		t.Errorf("Server = %q, want empty string", config.Server)
	}
	if config.Port != 0 {
		t.Errorf("Port = %d, want 0", config.Port)
	}
	if config.Username != "" {
		t.Errorf("Username = %q, want empty string", config.Username)
	}
	if config.Password != "" {
		t.Errorf("Password = %q, want empty string", config.Password)
	}
	if config.Name != "" {
		t.Errorf("Name = %q, want empty string", config.Name)
	}
}

func TestPostgresConfigJSONMarshaling(t *testing.T) {
	config := PostgresConfig{
		Config: Config{
			Server:   "localhost",
			Port:     5432,
			Username: "postgres",
			Password: "postgrespass",
			Name:     "postgresdb",
		},
		UseSSL: true,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal PostgresConfig to JSON: %v", err)
	}

	// Unmarshal back
	var decoded PostgresConfig
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal PostgresConfig from JSON: %v", err)
	}

	// Verify fields
	if decoded.Server != config.Server {
		t.Errorf("Server = %q, want %q", decoded.Server, config.Server)
	}
	if decoded.Port != config.Port {
		t.Errorf("Port = %d, want %d", decoded.Port, config.Port)
	}
	if decoded.Username != config.Username {
		t.Errorf("Username = %q, want %q", decoded.Username, config.Username)
	}
	if decoded.Password != config.Password {
		t.Errorf("Password = %q, want %q", decoded.Password, config.Password)
	}
	if decoded.Name != config.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, config.Name)
	}
	if decoded.UseSSL != config.UseSSL {
		t.Errorf("UseSSL = %v, want %v", decoded.UseSSL, config.UseSSL)
	}
}

func TestPostgresConfigYAMLMarshaling(t *testing.T) {
	config := PostgresConfig{
		Config: Config{
			Server:   "localhost",
			Port:     5432,
			Username: "postgres",
			Password: "postgrespass",
			Name:     "postgresdb",
		},
		UseSSL: false,
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal PostgresConfig to YAML: %v", err)
	}

	// Unmarshal back
	var decoded PostgresConfig
	if err := yaml.Unmarshal(yamlData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal PostgresConfig from YAML: %v", err)
	}

	// Verify fields
	if decoded.UseSSL != config.UseSSL {
		t.Errorf("UseSSL = %v, want %v", decoded.UseSSL, config.UseSSL)
	}
}

func TestPostgresConfigUseSSLDefault(t *testing.T) {
	config := PostgresConfig{}

	// Default value should be false
	if config.UseSSL != false {
		t.Errorf("UseSSL default = %v, want false", config.UseSSL)
	}
}

func TestPostgresConfigJSONFieldNames(t *testing.T) {
	jsonStr := `{"config":{"server":"pgserver","port":5432,"username":"pguser","password":"pgpass","name":"pgdb"},"use_ssl":true}`

	var config PostgresConfig
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if config.Server != "pgserver" {
		t.Errorf("Server = %q, want %q", config.Server, "pgserver")
	}
	if config.UseSSL != true {
		t.Errorf("UseSSL = %v, want true", config.UseSSL)
	}
}

func TestPostgresConfigYAMLFieldNames(t *testing.T) {
	yamlStr := `config:
  server: pgserver
  port: 5432
  username: pguser
  password: pgpass
  name: pgdb
use_ssl: true`

	var config PostgresConfig
	if err := yaml.Unmarshal([]byte(yamlStr), &config); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	if config.Server != "pgserver" {
		t.Errorf("Server = %q, want %q", config.Server, "pgserver")
	}
	if config.UseSSL != true {
		t.Errorf("UseSSL = %v, want true", config.UseSSL)
	}
}
