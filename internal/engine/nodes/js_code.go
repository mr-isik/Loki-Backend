package nodes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mr-isik/loki-backend/internal/domain"
	"github.com/mr-isik/loki-backend/internal/engine/docker"
)

type CodeJsNode struct{}

type codeJsData struct {
	Code  string                 `json:"code"`
	Input map[string]interface{} `json:"input"`
}

func (n *CodeJsNode) Execute(ctx context.Context, rawData []byte) (*domain.NodeResult, error) {
	var data codeJsData
	if err := json.Unmarshal(rawData, &data); err != nil {
		return &domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to parse input: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}

	// Initialize Docker runner
	runner, err := docker.NewRunner()
	if err != nil {
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Failed to initialize Docker runner: %v", err),
			OutputData: map[string]interface{}{
				"error": err.Error(),
			},
		}, nil
	}

	inputJSON, _ := json.Marshal(data.Input)
	codeJSON, _ := json.Marshal(data.Code) // Safely escape user code for JS eval

	wrapperScript := fmt.Sprintf(`
const input = %s;
const _logs = [];
const _originalLog = console.log;
console.log = function(...args) {
    _logs.push(args.map(a => typeof a === 'object' ? JSON.stringify(a) : String(a)).join(' '));
};

let _error = null;
let _result = null;

try {
    _result = eval(%s);
} catch (err) {
    _error = err.message || String(err);
}

process.stdout.write(JSON.stringify({ result: _result, logs: _logs, error: _error }));
`, string(inputJSON), string(codeJSON))

	outputStr, err := runner.RunContainer(ctx, docker.RunRequest{
		Image:   "node:alpine",
		Command: []string{"node", "-e", wrapperScript},
	})

	if err != nil {
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Node.js Execution Error: %v\nOutput: %s", err, outputStr),
			OutputData:      map[string]interface{}{"error": err.Error(), "logs": []string{outputStr}},
		}, nil
	}

	var parsedOutput struct {
		Result interface{} `json:"result"`
		Logs   []string    `json:"logs"`
		Error  *string     `json:"error"`
	}

	if err := json.Unmarshal([]byte(outputStr), &parsedOutput); err != nil {
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Failed to parse Node.js output: %v\nRaw Output: %s", err, outputStr),
			OutputData:      map[string]interface{}{"error": "invalid JSON output from container"},
		}, nil
	}

	if parsedOutput.Error != nil {
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("JS Execution Error: %s\nLogs: %v", *parsedOutput.Error, parsedOutput.Logs),
			OutputData:      map[string]interface{}{"error": *parsedOutput.Error},
		}, nil
	}

	// If output is not a map, wrap it
	var outputMap map[string]interface{}
	if m, ok := parsedOutput.Result.(map[string]interface{}); ok {
		outputMap = m
	} else {
		outputMap = map[string]interface{}{"result": parsedOutput.Result}
	}

	return &domain.NodeResult{
		Status:          "completed",
		TriggeredHandle: "output_success",
		Log:             fmt.Sprintf("JS Execution Success. Logs: %v", parsedOutput.Logs),
		OutputData:      outputMap,
	}, nil
}
