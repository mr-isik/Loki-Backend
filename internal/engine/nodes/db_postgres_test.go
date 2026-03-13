package nodes

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestDbPostgresNode_ParseData(t *testing.T) {
	
	validJSON := []byte(`{
		"host": "localhost",
		"port": 5432,
		"user": "testuser",
		"password": "testpassword",
		"dbname": "testdb",
		"query": "SELECT * FROM users WHERE age > $1 AND status = $2",
		"args": [18, "active"]
	}`)

	// Directly testing unmarshalling part of Execute (or mimicking it to verify structure mapping)
	var data dbPostgresData
	err := json.Unmarshal(validJSON, &data)

	if err != nil {
		t.Fatalf("Expected no error parsing valid JSON, got: %v", err)
	}

	if data.Host != "localhost" {
		t.Errorf("Expected Host 'localhost', got '%s'", data.Host)
	}

	if data.Port != 5432 {
		t.Errorf("Expected Port 5432, got %d", data.Port)
	}

	if data.Query != "SELECT * FROM users WHERE age > $1 AND status = $2" {
		t.Errorf("Expected specified query, got '%s'", data.Query)
	}

	expectedArgs := []interface{}{float64(18), "active"} // JSON numbers decode to float64 by default
	if !reflect.DeepEqual(data.Args, expectedArgs) {
		t.Errorf("Expected Args %#v, got %#v", expectedArgs, data.Args)
	}
}

func TestDbPostgresNode_FailedParse(t *testing.T) {
	node := &DbPostgresNode{}

	invalidJSON := []byte(`{ invalid }`)

	res, err := node.Execute(context.Background(), invalidJSON)

	if err == nil {
		t.Fatal("Expected error when parsing invalid JSON payload, got nil")
	}

	if res == nil {
		t.Fatal("Expected non-nil NodeResult even on failure")
	}

	if res.Status != "failed" {
		t.Errorf("Expected Status 'failed', got '%s'", res.Status)
	}
}
