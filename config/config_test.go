package config

import (
	"os"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	config := Default()

	// Test default values
	if config.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", config.Server.Port)
	}

	if config.Server.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got %s", config.Server.Host)
	}

	if config.Logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got %s", config.Logging.Level)
	}

	if config.AI.MaxThinkTime != 30*time.Second {
		t.Errorf("Expected default AI timeout 30s, got %v", config.AI.MaxThinkTime)
	}

	if config.AI.DefaultDifficulty != "medium" {
		t.Errorf("Expected default difficulty 'medium', got %s", config.AI.DefaultDifficulty)
	}

	if !config.Server.CORSEnabled {
		t.Error("Expected CORS to be enabled by default")
	}

	if !config.AI.EnableCaching {
		t.Error("Expected AI caching to be enabled by default")
	}
}

func TestConfigWithEnvironmentVariables(t *testing.T) {
	// Test with environment variables
	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(*Config) bool
	}{
		{
			name: "custom port",
			envVars: map[string]string{
				"CHESS_PORT": "9090",
			},
			validate: func(c *Config) bool { return c.Server.Port == 9090 },
		},
		{
			name: "custom host",
			envVars: map[string]string{
				"CHESS_HOST": "0.0.0.0",
			},
			validate: func(c *Config) bool { return c.Server.Host == "0.0.0.0" },
		},
		{
			name: "custom log level",
			envVars: map[string]string{
				"CHESS_LOG_LEVEL": "debug",
			},
			validate: func(c *Config) bool { return c.Logging.Level == "debug" },
		},
		{
			name: "custom AI timeout",
			envVars: map[string]string{
				"CHESS_AI_MAX_THINK_TIME": "60s",
			},
			validate: func(c *Config) bool { return c.AI.MaxThinkTime == 60*time.Second },
		},
		{
			name: "custom difficulty",
			envVars: map[string]string{
				"CHESS_AI_DEFAULT_DIFFICULTY": "hard",
			},
			validate: func(c *Config) bool { return c.AI.DefaultDifficulty == "hard" },
		},
		{
			name: "disable CORS",
			envVars: map[string]string{
				"CHESS_CORS_ENABLED": "false",
			},
			validate: func(c *Config) bool { return !c.Server.CORSEnabled },
		},
		{
			name: "enable LLM AI",
			envVars: map[string]string{
				"CHESS_LLMAI_ENABLED": "true",
			},
			validate: func(c *Config) bool { return c.LLMAI.Enabled },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			config := Default()
			if !tt.validate(config) {
				t.Errorf("Test %s failed: config does not match expectations", tt.name)
			}
		})
	}
}

func TestServerConfigAddress(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     int
		expected string
	}{
		{
			name:     "localhost with standard port",
			host:     "localhost",
			port:     8080,
			expected: "localhost:8080",
		},
		{
			name:     "all interfaces",
			host:     "0.0.0.0",
			port:     9000,
			expected: "0.0.0.0:9000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Default()
			config.Server.Host = tt.host
			config.Server.Port = tt.port

			addr := config.GetServerAddress()
			if addr != tt.expected {
				t.Errorf("Config.GetServerAddress() = %v, want %v", addr, tt.expected)
			}
		})
	}
}

func TestLLMAIProviderConfig(t *testing.T) {
	// Test LLM AI provider configuration with environment variables
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("OPENAI_MODEL", "gpt-4")
	defer func() {
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("OPENAI_MODEL")
	}()

	config := Default()

	if config.LLMAI.Providers["openai"].APIKey != "test-key" {
		t.Errorf("Expected OpenAI API key 'test-key', got %s", config.LLMAI.Providers["openai"].APIKey)
	}

	if config.LLMAI.Providers["openai"].Model != "gpt-4" {
		t.Errorf("Expected OpenAI model 'gpt-4', got %s", config.LLMAI.Providers["openai"].Model)
	}
}

func TestDatabaseConfig(t *testing.T) {
	config := Default()

	// Test default database configuration
	if config.Database.Driver == "" {
		t.Error("Database driver should have a default value")
	}

	if config.Database.MaxConnections <= 0 {
		t.Error("Database max connections should be positive")
	}

	if config.Database.ConnMaxLifetime <= 0 {
		t.Error("Database connection max lifetime should be positive")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  func() *Config
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  func() *Config { return Default() },
			wantErr: false,
		},
		{
			name: "invalid port (negative)",
			config: func() *Config {
				c := Default()
				c.Server.Port = -1
				return c
			},
			wantErr: true,
		},
		{
			name: "invalid port (too high)",
			config: func() *Config {
				c := Default()
				c.Server.Port = 65536
				return c
			},
			wantErr: true,
		},
		{
			name: "zero timeouts",
			config: func() *Config {
				c := Default()
				c.Server.ReadTimeout = 0
				return c
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.config()
			err := config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
