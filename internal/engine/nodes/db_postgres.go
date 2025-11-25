package nodes

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // Use pgx driver
	"github.com/mr-isik/loki-backend/internal/domain"
)

type DbPostgresNode struct{}

type dbPostgresData struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DbName   string `json:"dbname"`
	Query    string `json:"query"`
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

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		data.Host, data.Port, data.User, data.Password, data.DbName)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return &domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to connect to database: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}
	defer db.Close()

	// Simple query execution. For SELECT, we might want to return rows.
	// For INSERT/UPDATE/DELETE, we return result info.
	// This is a simplified implementation.

	rows, err := db.QueryContext(ctx, data.Query)
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
