package nodes

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // Use pgx driver
	"github.com/mr-isik/loki-backend/internal/domain"
)

type DbPostgresNode struct{}

// dbPool caches connection pools based on their secure DSN representation
// avoiding repeating the SSL handshake and connection logic overhead per node execution
var dbPool sync.Map

type dbPostgresData struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string        `json:"user"`
	Password string        `json:"password"`
	DbName   string        `json:"dbname"`
	Query    string        `json:"query"`
	Args     []interface{} `json:"args"` // Used for parameterization preventing SQL injections
}

func (n *DbPostgresNode) Execute(ctx context.Context, rawData []byte) (*domain.NodeResult, error) {
	var data dbPostgresData
	if err := json.Unmarshal(rawData, &data); err != nil {
		return &domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to parse input: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}


	// Build a secure DSN instead of raw string formatting
	dbURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(data.User, data.Password),
		Host:   fmt.Sprintf("%s:%d", data.Host, data.Port),
		Path:   data.DbName,
	}
	q := dbURL.Query()
	q.Set("sslmode", "disable") // Can be configurable via data in future
	dbURL.RawQuery = q.Encode()
	dsn := dbURL.String()

	var db *sql.DB

	// Check connection pool cache
	if pooledDb, ok := dbPool.Load(dsn); ok {
		db = pooledDb.(*sql.DB)
	} else {
		newDb, err := sql.Open("pgx", dsn)
		if err != nil {
			return &domain.NodeResult{
				Status:     "failed",
				Log:        fmt.Sprintf("Failed to initialize database pool: %v", err),
				OutputData: map[string]interface{}{"error": err.Error()},
			}, err
		}

		// Configure optimal pool limits
		newDb.SetMaxOpenConns(25)                 // Max parallel conns
		newDb.SetMaxIdleConns(5)                  // Keep alive idle conns
		newDb.SetConnMaxLifetime(5 * time.Minute) // Maximum duration a connection is kept alive

		dbPool.Store(dsn, newDb)
		db = newDb
	}

	// Simple query execution. For SELECT, we might want to return rows.
	// For INSERT/UPDATE/DELETE, we return result info.
	// This is a simplified implementation.

	var rows *sql.Rows
	var err error

	// Pass arguments to parameters if provided, enabling SQL injection protection
	if len(data.Args) > 0 {
		rows, err = db.QueryContext(ctx, data.Query, data.Args...)
	} else {
		rows, err = db.QueryContext(ctx, data.Query)
	}

	if err != nil {
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Query failed: %v", err),
			OutputData:      map[string]interface{}{"error": err.Error()},
		}, nil
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	var results []map[string]interface{}

	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columnsPtrs := make([]interface{}, len(columns))
		columnValues := make([]interface{}, len(columns))
		for i := range columnValues {
			columnsPtrs[i] = &columnValues[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnsPtrs...); err != nil {
			continue
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]interface{})
		for i, colName := range columns {
			val := columnValues[i]
			b, ok := val.([]byte)
			if ok {
				m[colName] = string(b)
			} else {
				m[colName] = val
			}
		}
		results = append(results, m)
	}

	return &domain.NodeResult{
		Status:          "completed",
		TriggeredHandle: "output_success",
		Log:             fmt.Sprintf("Query executed successfully. Rows returned: %d", len(results)),
		OutputData: map[string]interface{}{
			"rows": results,
		},
	}, nil
}
