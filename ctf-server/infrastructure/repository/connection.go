package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	Host       string
	Port       string
	User       string
	Password   string
	Database   string
	SchemaPath string
}

func NewConfigFromEnv() *Config {
	config := &Config{
		Host:       getEnv("DB_HOST", "localhost"),
		Port:       getEnv("DB_PORT", "3306"),
		User:       getEnv("DB_USER", "root"),
		Password:   getEnv("DB_PASSWORD", "password"),
		Database:   getEnv("DB_NAME", "ctf_server_db"),
		SchemaPath: getEnv("SCHEMA_PATH", "../migration/ctf_server_schema.sql"),
	}
	return config
}

func Connect(config *Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Connected to database: %s", config.Database)

	return db, nil
}

func InitSchema(ctx context.Context, db *sql.DB, schemaPath string) error {
	schemaSQL, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	statements := splitSQL(string(schemaSQL))
	for _, stmt := range statements {
		if stmt == "" {
			continue
		}

		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	log.Printf("Schema initialized from: %s", schemaPath)
	return nil
}

func splitSQL(sql string) []string {
	var statements []string
	var current string

	for _, line := range strings.Split(sql, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		current += line + "\n"

		if strings.HasSuffix(line, ";") {
			statements = append(statements, current)
			current = ""
		}
	}

	if current != "" {
		statements = append(statements, current)
	}

	return statements
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
