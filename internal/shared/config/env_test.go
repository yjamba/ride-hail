package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFile_FileNotExist(t *testing.T) {
	err := LoadEnvFile("nonexistent.env")
	if err != nil {
		t.Fatalf("expected nil for nonexistent file, got %v", err)
	}
}

func TestLoadEnvFile_BasicParsing(t *testing.T) {
	// Create temp .env file
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := `# Comment line
KEY1=value1
KEY2=value2
KEY3="quoted value"
KEY4='single quoted'
EMPTY_LINE_ABOVE=yes

NO_EQUALS_LINE
  SPACES_AROUND  =  trimmed  
`
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Clear any existing values
	os.Unsetenv("KEY1")
	os.Unsetenv("KEY2")
	os.Unsetenv("KEY3")
	os.Unsetenv("KEY4")
	os.Unsetenv("EMPTY_LINE_ABOVE")
	os.Unsetenv("SPACES_AROUND")

	err := LoadEnvFile(envFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"KEY1", "value1"},
		{"KEY2", "value2"},
		{"KEY3", "quoted value"},
		{"KEY4", "single quoted"},
		{"EMPTY_LINE_ABOVE", "yes"},
		{"SPACES_AROUND", "trimmed"},
	}

	for _, tc := range tests {
		got := os.Getenv(tc.key)
		if got != tc.expected {
			t.Errorf("key %s: expected %q, got %q", tc.key, tc.expected, got)
		}
	}

	// Cleanup
	for _, tc := range tests {
		os.Unsetenv(tc.key)
	}
}

func TestLoadEnvFile_NoOverwrite(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := `EXISTING_KEY=new_value`
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Set existing value
	os.Setenv("EXISTING_KEY", "original_value")
	defer os.Unsetenv("EXISTING_KEY")

	err := LoadEnvFile(envFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := os.Getenv("EXISTING_KEY")
	if got != "original_value" {
		t.Errorf("expected original_value, got %q (should not overwrite)", got)
	}
}

func TestLoadEnv_UsesDefaultPath(t *testing.T) {
	// LoadEnv should not fail if .env doesn't exist
	err := LoadEnv()
	if err != nil {
		t.Fatalf("LoadEnv should return nil when .env doesn't exist, got %v", err)
	}
}
