// Package config provides configuration management for the chess engine.
// It includes server settings, AI configuration, and other application settings.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config represents the application configuration.
type Config struct {
	Server   ServerConfig   `json:"server"`
	AI       AIConfig       `json:"ai"`
	LLMAI    LLMAIConfig    `json:"llm_ai"`
	Logging  LoggingConfig  `json:"logging"`
	Database DatabaseConfig `json:"database"`
}

// ServerConfig contains HTTP server configuration.
type ServerConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	IdleTimeout     time.Duration `json:"idle_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
	CORSEnabled     bool          `json:"cors_enabled"`
	AllowedOrigins  []string      `json:"allowed_origins"`
}

// AIConfig contains AI engine configuration.
type AIConfig struct {
	DefaultDifficulty string        `json:"default_difficulty"`
	MaxThinkTime      time.Duration `json:"max_think_time"`
	EnableCaching     bool          `json:"enable_caching"`
	CacheSize         int           `json:"cache_size"`
}

// LLMAIConfig contains LLM AI provider configuration.
type LLMAIConfig struct {
	Enabled         bool                         `json:"enabled"`
	DefaultProvider string                       `json:"default_provider"`
	ChatEnabled     bool                         `json:"chat_enabled"`
	Providers       map[string]LLMProviderConfig `json:"providers"`
}

// LLMProviderConfig contains configuration for a specific LLM provider.
type LLMProviderConfig struct {
	APIKey      string `json:"api_key"`
	Model       string `json:"model"`
	Endpoint    string `json:"endpoint"`
	Personality string `json:"personality"`
}

// LoggingConfig contains logging configuration.
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	OutputPath string `json:"output_path"`
	ErrorPath  string `json:"error_path"`
}

// DatabaseConfig contains database configuration.
type DatabaseConfig struct {
	Driver           string        `json:"driver"`
	ConnectionString string        `json:"connection_string"`
	MaxConnections   int           `json:"max_connections"`
	ConnMaxLifetime  time.Duration `json:"conn_max_lifetime"`
	MigrationsPath   string        `json:"migrations_path"`
}

// Default returns a default configuration.
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host:            getEnvString("CHESS_HOST", "localhost"),
			Port:            getEnvInt("CHESS_PORT", 8080),
			ReadTimeout:     getEnvDuration("CHESS_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getEnvDuration("CHESS_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:     getEnvDuration("CHESS_IDLE_TIMEOUT", 120*time.Second),
			ShutdownTimeout: getEnvDuration("CHESS_SHUTDOWN_TIMEOUT", 10*time.Second),
			CORSEnabled:     getEnvBool("CHESS_CORS_ENABLED", true),
			AllowedOrigins:  getEnvStringSlice("CHESS_ALLOWED_ORIGINS", []string{"*"}),
		},
		AI: AIConfig{
			DefaultDifficulty: getEnvString("CHESS_AI_DEFAULT_DIFFICULTY", "medium"),
			MaxThinkTime:      getEnvDuration("CHESS_AI_MAX_THINK_TIME", 30*time.Second),
			EnableCaching:     getEnvBool("CHESS_AI_ENABLE_CACHING", true),
			CacheSize:         getEnvInt("CHESS_AI_CACHE_SIZE", 1000),
		},
		LLMAI: LLMAIConfig{
			Enabled:         getEnvBool("CHESS_LLMAI_ENABLED", false),
			DefaultProvider: getEnvString("CHESS_LLMAI_PROVIDER", "openai"),
			ChatEnabled:     getEnvBool("CHESS_LLMAI_CHAT", true),
			Providers: map[string]LLMProviderConfig{
				"openai": {
					APIKey:      getEnvString("OPENAI_API_KEY", ""),
					Model:       getEnvString("OPENAI_MODEL", "gpt-3.5-turbo"),
					Endpoint:    getEnvString("OPENAI_ENDPOINT", "https://api.openai.com/v1/chat/completions"),
					Personality: getEnvString("OPENAI_PERSONALITY", "a friendly but competitive chess master"),
				},
				"anthropic": {
					APIKey:      getEnvString("ANTHROPIC_API_KEY", ""),
					Model:       getEnvString("ANTHROPIC_MODEL", "claude-3-haiku-20240307"),
					Endpoint:    getEnvString("ANTHROPIC_ENDPOINT", "https://api.anthropic.com/v1/messages"),
					Personality: getEnvString("ANTHROPIC_PERSONALITY", "a thoughtful and analytical chess strategist"),
				},
				"gemini": {
					APIKey:      getEnvString("GEMINI_API_KEY", ""),
					Model:       getEnvString("GEMINI_MODEL", "gemini-1.5-flash"),
					Endpoint:    getEnvString("GEMINI_ENDPOINT", "https://generativelanguage.googleapis.com/v1beta/models"),
					Personality: getEnvString("GEMINI_PERSONALITY", "a creative and intuitive chess player"),
				},
				"xai": {
					APIKey:      getEnvString("XAI_API_KEY", ""),
					Model:       getEnvString("XAI_MODEL", "grok-beta"),
					Endpoint:    getEnvString("XAI_ENDPOINT", "https://api.x.ai/v1/chat/completions"),
					Personality: getEnvString("XAI_PERSONALITY", "a witty and clever chess opponent"),
				},
				"deepseek": {
					APIKey:      getEnvString("DEEPSEEK_API_KEY", ""),
					Model:       getEnvString("DEEPSEEK_MODEL", "deepseek-chat"),
					Endpoint:    getEnvString("DEEPSEEK_ENDPOINT", "https://api.deepseek.com/v1/chat/completions"),
					Personality: getEnvString("DEEPSEEK_PERSONALITY", "a deep-thinking and methodical chess AI"),
				},
			},
		},
		Logging: LoggingConfig{
			Level:      getEnvString("CHESS_LOG_LEVEL", "info"),
			Format:     getEnvString("CHESS_LOG_FORMAT", "json"),
			OutputPath: getEnvString("CHESS_LOG_OUTPUT_PATH", "stdout"),
			ErrorPath:  getEnvString("CHESS_LOG_ERROR_PATH", "stderr"),
		},
		Database: DatabaseConfig{
			Driver:           getEnvString("CHESS_DB_DRIVER", "sqlite3"),
			ConnectionString: getEnvString("CHESS_DB_CONNECTION_STRING", "./chess.db"),
			MaxConnections:   getEnvInt("CHESS_DB_MAX_CONNECTIONS", 10),
			ConnMaxLifetime:  getEnvDuration("CHESS_DB_CONN_MAX_LIFETIME", 1*time.Hour),
			MigrationsPath:   getEnvString("CHESS_DB_MIGRATIONS_PATH", "./migrations"),
		},
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	// Validate server configuration
	if c.Server.Port < 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be between 0 and 65535)", c.Server.Port)
	}

	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("invalid server read timeout: %v (must be positive)", c.Server.ReadTimeout)
	}

	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("invalid server write timeout: %v (must be positive)", c.Server.WriteTimeout)
	}

	// Validate AI configuration
	if c.AI.MaxThinkTime <= 0 {
		return fmt.Errorf("invalid AI max think time: %v (must be positive)", c.AI.MaxThinkTime)
	}

	// Validate LLMAI configuration
	if c.LLMAI.Enabled {
		if c.LLMAI.DefaultProvider == "" {
			return fmt.Errorf("LLMAI is enabled but no default provider is set")
		}
	}

	return nil
}

// GetServerAddress returns the full server address.
func (c *Config) GetServerAddress() string {
	return c.Server.Host + ":" + strconv.Itoa(c.Server.Port)
}

// GetLLMProviderConfig returns the configuration for a specific LLM provider.
func (c *Config) GetLLMProviderConfig(provider string) (LLMProviderConfig, bool) {
	cfg, exists := c.LLMAI.Providers[provider]
	return cfg, exists
}

// HasValidLLMProvider checks if the specified provider has a valid API key.
func (c *Config) HasValidLLMProvider(provider string) bool {
	if !c.LLMAI.Enabled {
		return false
	}

	cfg, exists := c.GetLLMProviderConfig(provider)
	if !exists {
		return false
	}

	// DeepSeek might work without API key in some configurations
	if provider == "deepseek" {
		return true
	}

	return cfg.APIKey != ""
}

// GetAvailableLLMProviders returns a list of providers with valid API keys.
func (c *Config) GetAvailableLLMProviders() []string {
	if !c.LLMAI.Enabled {
		return []string{}
	}

	var providers []string
	for name := range c.LLMAI.Providers {
		if c.HasValidLLMProvider(name) {
			providers = append(providers, name)
		}
	}
	return providers
}

// Helper functions for environment variable parsing

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		// For a production app, you might want more sophisticated parsing
		return []string{value}
	}
	return defaultValue
}
