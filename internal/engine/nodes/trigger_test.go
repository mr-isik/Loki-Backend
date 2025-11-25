package nodes

import (
	"context"
	"testing"
)

func TestWebhookNode_Execute(t *testing.T) {
	node := &WebhookNode{}
	ctx := context.Background()

	input := []byte(`{"foo":"bar"}`)
	result, err := node.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Status != "completed" {
		t.Errorf("Expected status completed, got %s", result.Status)
	}

	payload := result.OutputData["payload"].(string)
	if payload != `{"foo":"bar"}` {
		t.Errorf("Expected payload %s, got %s", `{"foo":"bar"}`, payload)
	}
}

func TestCronNode_Execute(t *testing.T) {
	node := &CronNode{}
	ctx := context.Background()

	result, err := node.Execute(ctx, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Status != "completed" {
		t.Errorf("Expected status completed, got %s", result.Status)
	}

	if _, ok := result.OutputData["timestamp"]; !ok {
		t.Error("Expected timestamp in output")
	}
}
