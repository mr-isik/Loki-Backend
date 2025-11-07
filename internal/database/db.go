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

// RunMigrations executes all database migrations
func (db *Database) RunMigrations(ctx context.Context) error {
	log.Println("🔄 Running database migrations...")

	migrations := []struct {
		name string
		sql  string
	}{
		{
			name: "001_create_users_table",
			sql: `
				-- Create users table
				CREATE TABLE IF NOT EXISTS users (
					id UUID PRIMARY KEY,
					email VARCHAR(255) UNIQUE NOT NULL,
					name VARCHAR(100) NOT NULL,
					password VARCHAR(255) NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
					deleted_at TIMESTAMP
				);

				-- Create indexes
				CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE deleted_at IS NULL;
				CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);
				CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
			`,
		},
		{
			name: "002_create_workspaces_table",
			sql: `
				-- Create workspaces table
				CREATE TABLE IF NOT EXISTS workspaces (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					name VARCHAR(255) NOT NULL,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				-- Create indexes
				CREATE INDEX IF NOT EXISTS idx_workspaces_owner_user_id ON workspaces(owner_user_id);
				CREATE INDEX IF NOT EXISTS idx_workspaces_created_at ON workspaces(created_at DESC);
			`,
		},
		{
			name: "003_create_workflows_table",
			sql: `
				-- Create workflow_status enum
				DO $$ BEGIN
					CREATE TYPE workflow_status AS ENUM ('draft', 'published', 'archived');
				EXCEPTION
					WHEN duplicate_object THEN null;
				END $$;

				-- Create workflows table
				CREATE TABLE IF NOT EXISTS workflows (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
					title VARCHAR(255) NOT NULL DEFAULT 'Untitled Workflow',
					status workflow_status NOT NULL DEFAULT 'draft',
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				-- Create indexes
				CREATE INDEX IF NOT EXISTS idx_workflows_workspace_id ON workflows(workspace_id);
				CREATE INDEX IF NOT EXISTS idx_workflows_status ON workflows(status);
				CREATE INDEX IF NOT EXISTS idx_workflows_updated_at ON workflows(updated_at DESC);
				CREATE INDEX IF NOT EXISTS idx_workflows_created_at ON workflows(created_at DESC);
			`,
		},
		{
			name: "004_create_node_templates_table",
			sql: `
				-- Create node_templates table
				CREATE TABLE IF NOT EXISTS node_templates (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					name VARCHAR(255) NOT NULL,
					description TEXT,
					type_key VARCHAR(100) NOT NULL UNIQUE,
					category VARCHAR(100) NOT NULL,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				-- Create indexes
				CREATE INDEX IF NOT EXISTS idx_node_templates_type_key ON node_templates(type_key);
				CREATE INDEX IF NOT EXISTS idx_node_templates_category ON node_templates(category);
				
				-- Insert default node templates
				INSERT INTO node_templates (name, description, type_key, category) VALUES
					('HTTP Request', 'Make HTTP requests to external APIs', 'http_request', 'integration'),
					('Shell Command', 'Execute shell commands', 'shell_command', 'utility'),
					('Condition', 'Conditional branching based on data', 'condition', 'control'),
					('Loop', 'Iterate over data collections', 'loop', 'control')
				ON CONFLICT (type_key) DO NOTHING;
			`,
		},
		{
			name: "004_create_workflow_edges_table",
			sql: `
				-- Create workflow_edges table
				CREATE TABLE IF NOT EXISTS workflow_edges (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
					source_node_id UUID NOT NULL,
					target_node_id UUID NOT NULL,
					source_handle VARCHAR(255) NOT NULL,
					target_handle VARCHAR(255) NOT NULL,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				-- Create indexes for workflow_edges
				CREATE INDEX IF NOT EXISTS idx_workflow_edges_workflow_id ON workflow_edges(workflow_id);
				CREATE INDEX IF NOT EXISTS idx_workflow_edges_source_node ON workflow_edges(source_node_id);
				CREATE INDEX IF NOT EXISTS idx_workflow_edges_target_node ON workflow_edges(target_node_id);
			`,
		},
		{
			name: "005_create_workflow_nodes_table",
			sql: `
				-- Create workflow_nodes table
				CREATE TABLE IF NOT EXISTS workflow_nodes (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
					template_id UUID NOT NULL REFERENCES node_templates(id) ON DELETE RESTRICT,
					position_x FLOAT NOT NULL DEFAULT 0,
					position_y FLOAT NOT NULL DEFAULT 0,
					data JSONB,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				-- Create indexes for workflow_nodes
				CREATE INDEX IF NOT EXISTS idx_workflow_nodes_workflow_id ON workflow_nodes(workflow_id);
				CREATE INDEX IF NOT EXISTS idx_workflow_nodes_template_id ON workflow_nodes(template_id);
			`,
		},
	}

	// Execute migrations in order
	for _, migration := range migrations {
		log.Printf("  → Running migration: %s", migration.name)
		
		if _, err := db.Pool.Exec(ctx, migration.sql); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.name, err)
		}
		
		log.Printf("  ✅ Migration %s completed", migration.name)
	}

	log.Println("✅ All migrations completed successfully")
	return nil
}