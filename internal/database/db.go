package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Database wraps the pgxpool connection
type Database struct {
	Pool *pgxpool.Pool
}

// NewConfig creates a new database configuration
func NewConfig(host, port, user, password, dbname string) *Config {
	return &Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
		SSLMode:  "disable",
	}
}

// ConnectionString returns the PostgreSQL connection string
func (c *Config) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// NewDatabase creates a new database connection pool
func NewDatabase(config *Config) (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(config.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	// Configure pool settings
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("✅ Database connection established successfully")

	return &Database{Pool: pool}, nil
}

// Close closes the database connection pool
func (db *Database) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		log.Println("Database connection closed")
	}
}

// Health checks if database is healthy
func (db *Database) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	return db.Pool.Ping(ctx)
}