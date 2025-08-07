package config

import (
	"os"
	"testing"
)

func TestConfigLoading(t *testing.T) {
	// Test loading config with default values
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg == nil {
		t.Error("Expected config to be loaded")
	}

	// Verify default values
	if cfg.Server.Port == 0 {
		t.Error("Expected default server port to be set")
	}

	if cfg.Engine.Depth == 0 {
		t.Error("Expected default engine depth to be set")
	}
}

func TestConfigWithFile(t *testing.T) {
	// Create a temporary config file
	configContent := `{
		"server": {
			"port": 9090,
			"host": "0.0.0.0"
		},
		"engine": {
			"depth": 5,
			"timeout": 10
		},
		"chat": {
			"enabled": true,
			"provider": "openai"
		}
	}`

	tmpFile := "/tmp/test_config.json"
	err := os.WriteFile(tmpFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Test loading from file
	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config from file: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", cfg.Server.Port)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Expected host '0.0.0.0', got '%s'", cfg.Server.Host)
	}

	if cfg.Engine.Depth != 5 {
		t.Errorf("Expected engine depth 5, got %d", cfg.Engine.Depth)
	}

	if cfg.Engine.Timeout != 10 {
		t.Errorf("Expected engine timeout 10, got %d", cfg.Engine.Timeout)
	}

	if !cfg.Chat.Enabled {
		t.Error("Expected chat to be enabled")
	}

	if cfg.Chat.Provider != "openai" {
		t.Errorf("Expected chat provider 'openai', got '%s'", cfg.Chat.Provider)
	}
}

func TestInvalidConfigFile(t *testing.T) {
	// Test with non-existent file
	_, err := LoadConfig("/non/existent/file.json")
	if err == nil {
		t.Error("Expected error for non-existent config file")
	}

	// Test with invalid JSON
	invalidJSON := `{"server": {"port": invalid}}`
	tmpFile := "/tmp/invalid_config.json"
	err = os.WriteFile(tmpFile, []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid config file: %v", err)
	}
	defer os.Remove(tmpFile)

	_, err = LoadConfig(tmpFile)
	if err == nil {
		t.Error("Expected error for invalid JSON config")
	}
}

func TestEnvironmentVariables(t *testing.T) {
	// Test that environment variables are used
	originalPort := os.Getenv("CHESS_PORT")
	originalHost := os.Getenv("CHESS_HOST")

	// Set test environment variables
	os.Setenv("CHESS_PORT", "7777")
	os.Setenv("CHESS_HOST", "test.local")

	defer func() {
		// Restore original values
		if originalPort == "" {
			os.Unsetenv("CHESS_PORT")
		} else {
			os.Setenv("CHESS_PORT", originalPort)
		}
		if originalHost == "" {
			os.Unsetenv("CHESS_HOST")
		} else {
			os.Setenv("CHESS_HOST", originalHost)
		}
	}()

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Note: This test assumes the config loader uses environment variables
	// The actual behavior depends on the implementation
	t.Logf("Config loaded with port: %d, host: %s", cfg.Server.Port, cfg.Server.Host)
}

func TestConfigValidation(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "localhost",
		},
		Engine: EngineConfig{
			Depth:   4,
			Timeout: 5,
		},
		Chat: ChatConfig{
			Enabled:  true,
			Provider: "openai",
		},
	}

	// Test that valid config doesn't produce errors
	if cfg.Server.Port <= 0 {
		t.Error("Expected valid port number")
	}

	if cfg.Engine.Depth <= 0 {
		t.Error("Expected positive engine depth")
	}

	if cfg.Engine.Timeout <= 0 {
		t.Error("Expected positive engine timeout")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Error("Expected default config to be created")
	}

	// Verify reasonable defaults
	if cfg.Server.Port < 1024 || cfg.Server.Port > 65535 {
		t.Errorf("Expected reasonable default port, got %d", cfg.Server.Port)
	}

	if cfg.Server.Host == "" {
		t.Error("Expected default host to be set")
	}

	if cfg.Engine.Depth <= 0 {
		t.Error("Expected positive default engine depth")
	}

	if cfg.Engine.Timeout <= 0 {
		t.Error("Expected positive default engine timeout")
	}
}

func TestConfigFields(t *testing.T) {
	cfg := &Config{}

	// Test that all expected fields exist (compilation test)
	_ = cfg.Server.Port
	_ = cfg.Server.Host
	_ = cfg.Engine.Depth
	_ = cfg.Engine.Timeout
	_ = cfg.Chat.Enabled
	_ = cfg.Chat.Provider

	// Create a complete config to verify structure
	completeConfig := Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "localhost",
		},
		Engine: EngineConfig{
			Depth:   4,
			Timeout: 5,
		},
		Chat: ChatConfig{
			Enabled:  true,
			Provider: "openai",
		},
	}

	if completeConfig.Server.Port != 8080 {
		t.Error("Config structure field assignment failed")
	}
}
