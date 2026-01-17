package postgres

import "strings"

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewDBConfig(localhost, port, user, password, dbName, sslMode string) *DBConfig {
	// Normalize SSLMode to handle common mistakes like "disabled" → "disable"
	normalizedSSLMode := normalizeSSLMode(sslMode)

	return &DBConfig{
		Host:     localhost,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
		SSLMode:  normalizedSSLMode,
	}
}

// normalizeSSLMode converts common invalid sslmode values to valid ones
func normalizeSSLMode(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))

	// Map of invalid values to correct ones
	corrections := map[string]string{
		"disabled": "disable",
		"enabled":  "require",
		"true":     "require",
		"false":    "disable",
	}

	if corrected, exists := corrections[mode]; exists {
		return corrected
	}

	// Valid modes: disable, allow, prefer, require, verify-ca, verify-full
	validModes := map[string]bool{
		"disable":     true,
		"allow":       true,
		"prefer":      true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}

	if validModes[mode] {
		return mode
	}

	// Default to "disable" if invalid mode is provided
	return "disable"
}
