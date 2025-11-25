package nodes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttpRequestNode_Execute(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/success" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "success"}`))
			return
		}
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal server error"}`))
			return
		}
		if r.URL.Path == "/echo" {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			json.NewEncoder(w).Encode(body)
			return
		}
	}))
	defer server.Close()

	node := &HttpRequestNode{}
	ctx := context.Background()

	t.Run("Success Request", func(t *testing.T) {
		input := map[string]interface{}{
			"url":    server.URL + "/success",
			"method": "GET",
		}
		inputBytes, _ := json.Marshal(input)

		result, err := node.Execute(ctx, inputBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Status != "completed" {
			t.Errorf("Expected status completed, got %s", result.Status)
		}
		if result.TriggeredHandle != "output_success" {
			t.Errorf("Expected handle output_success, got %s", result.TriggeredHandle)
		}

		outputData := result.OutputData
		if outputData["status"] != 200 {
			t.Errorf("Expected status 200, got %v", outputData["status"])
		}
	})

	t.Run("Error Request", func(t *testing.T) {
		input := map[string]interface{}{
			"url":    server.URL + "/error",
			"method": "GET",
		}
		inputBytes, _ := json.Marshal(input)

		result, err := node.Execute(ctx, inputBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Status != "completed" {
			t.Errorf("Expected status completed, got %s", result.Status)
		}

		outputData := result.OutputData
		if outputData["status"] != 500 {
			t.Errorf("Expected status 500, got %v", outputData["status"])
		}
	})

	t.Run("Post Request with Body", func(t *testing.T) {
		input := map[string]interface{}{
			"url":    server.URL + "/echo",
			"method": "POST",
			"body": map[string]interface{}{
				"foo": "bar",
			},
		}
		inputBytes, _ := json.Marshal(input)

		result, err := node.Execute(ctx, inputBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		outputData := result.OutputData
		body := outputData["body"].(map[string]interface{})
		if body["foo"] != "bar" {
			t.Errorf("Expected body foo=bar, got %v", body["foo"])
		}
	})

	t.Run("Invalid URL", func(t *testing.T) {
		input := map[string]interface{}{
			"url": "",
		}
		inputBytes, _ := json.Marshal(input)

		_, err := node.Execute(ctx, inputBytes)
		if err == nil {
			t.Fatal("Expected error for empty URL, got nil")
		}
	})
}
