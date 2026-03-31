package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Security SecurityConfig `yaml:"security"`
	Database DatabaseConfig `yaml:"database"`
	CORS     CORSConfig     `yaml:"cors"`
	Redis    RedisConfig    `yaml:"redis"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func (c ServerConfig) Address() string {
	return c.Host + ":" + c.Port
}

type SecurityConfig struct {
	AdminSessionSecret string `yaml:"admin_session_secret"`
	EncryptionKey      string `yaml:"encryption_key"`
	AdminPasswordHash  string `yaml:"admin_password_hash"`
	AdminPasswordPlain string `yaml:"admin_password_plain"`
}

type DatabaseConfig struct {
	Driver          string `yaml:"driver"`
	DSN             string `yaml:"dsn"`
	InitSQLPath     string `yaml:"init_sql_path"`
	MaxOpenConns    int32  `yaml:"max_open_conns"`
	MinIdleConns    int32  `yaml:"min_idle_conns"`
	MaxConnLifetime int    `yaml:"max_conn_lifetime_minutes"`
	MaxConnIdleTime int    `yaml:"max_conn_idle_minutes"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func Load() (Config, error) {
	path := os.Getenv("LOCALAIHUB_CONFIG")
	if path == "" {
		path = "configs/config.local.yaml"
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = "configs/config.example.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	overrideFromEnv(&cfg)
	applyDefaults(&cfg)

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if c.Server.Host == "" {
		return fmt.Errorf("server.host is required")
	}
	if c.Server.Port == "" {
		return fmt.Errorf("server.port is required")
	}
	if c.Database.Driver == "" {
		return fmt.Errorf("database.driver is required")
	}
	if c.Database.DSN == "" {
		return fmt.Errorf("database.dsn is required")
	}
	if c.Security.AdminSessionSecret == "" {
		return fmt.Errorf("security.admin_session_secret is required")
	}
	if c.Security.EncryptionKey == "" {
		return fmt.Errorf("security.encryption_key is required")
	}
	return nil
}

func overrideFromEnv(cfg *Config) {
	setString(&cfg.Server.Host, "LOCALAIHUB_SERVER_HOST")
	setString(&cfg.Server.Port, "LOCALAIHUB_SERVER_PORT")
	setString(&cfg.Security.AdminSessionSecret, "LOCALAIHUB_ADMIN_SESSION_SECRET")
	setString(&cfg.Security.EncryptionKey, "LOCALAIHUB_ENCRYPTION_KEY")
	setString(&cfg.Security.AdminPasswordHash, "LOCALAIHUB_ADMIN_PASSWORD_HASH")
	setString(&cfg.Security.AdminPasswordPlain, "LOCALAIHUB_ADMIN_PASSWORD_PLAIN")
	setString(&cfg.Database.Driver, "LOCALAIHUB_DB_DRIVER")
	setString(&cfg.Database.DSN, "LOCALAIHUB_DB_DSN")
	setString(&cfg.Database.InitSQLPath, "LOCALAIHUB_DB_INIT_SQL_PATH")
	setInt32(&cfg.Database.MaxOpenConns, "LOCALAIHUB_DB_MAX_OPEN_CONNS")
	setInt32(&cfg.Database.MinIdleConns, "LOCALAIHUB_DB_MIN_IDLE_CONNS")
	setInt(&cfg.Database.MaxConnLifetime, "LOCALAIHUB_DB_MAX_CONN_LIFETIME_MINUTES")
	setInt(&cfg.Database.MaxConnIdleTime, "LOCALAIHUB_DB_MAX_CONN_IDLE_MINUTES")
}

func applyDefaults(cfg *Config) {
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 10
	}
	if cfg.Database.InitSQLPath == "" {
		cfg.Database.InitSQLPath = "sql/init.sql"
	}
	if cfg.Database.MinIdleConns == 0 {
		cfg.Database.MinIdleConns = 2
	}
	if len(cfg.CORS.AllowedOrigins) == 0 {
		cfg.CORS.AllowedOrigins = []string{"http://127.0.0.1:3000", "http://localhost:3000"}
	}
	if cfg.Database.MaxConnLifetime == 0 {
		cfg.Database.MaxConnLifetime = 30
	}
	if cfg.Database.MaxConnIdleTime == 0 {
		cfg.Database.MaxConnIdleTime = 10
	}
}

func setString(target *string, envKey string) {
	if value := os.Getenv(envKey); value != "" {
		*target = value
	}
}

func setInt(target *int, envKey string) {
	if value := os.Getenv(envKey); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil {
			*target = parsed
		}
	}
}

func setInt32(target *int32, envKey string) {
	if value := os.Getenv(envKey); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil {
			*target = int32(parsed)
		}
	}
}
