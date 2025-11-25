package nodes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSlackNode_Execute(t *testing.T) {
	// Mock Slack API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["text"] != "hello slack" {
			t.Errorf("Expected message 'hello slack', got %v", body["text"])
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	node := &SlackNode{}
	ctx := context.Background()

	input := map[string]interface{}{
		"webhook_url": server.URL,
		"message":     "hello slack",
	}
	inputBytes, _ := json.Marshal(input)

	result, err := node.Execute(ctx, inputBytes)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Status != "completed" {
		t.Errorf("Expected status completed, got %s", result.Status)
	}
}

func TestEmailSmtpNode_Execute(t *testing.T) {
	// We can't easily mock SMTP without a library or a complex mock server.
	// For now, we'll just test input validation or basic structure.
	// Or we can skip actual sending if we mock net/smtp (hard in Go without interfaces).
	// Let's just test that it fails gracefully with invalid host.

	node := &EmailSmtpNode{}
	ctx := context.Background()

	input := map[string]interface{}{
		"host": "invalid-host",
		"port": 25,
	}
	inputBytes, _ := json.Marshal(input)

	result, err := node.Execute(ctx, inputBytes)
	// It should return success=false (status=failed) but err might be nil as per our implementation
	// Wait, our implementation returns nil error on send failure, but sets status to failed.

	if err != nil {
		// It's okay if it returns error too, but let's check result
	}

	if result.Status != "failed" {
		t.Errorf("Expected status failed for invalid host, got %s", result.Status)
	}
}
