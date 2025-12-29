package postgres

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewDBConfig(localhost, port, user, password, dbName, sslMode string) *DBConfig {
	return &DBConfig{
		Host:     localhost,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
		SSLMode:  sslMode,
	}
}
