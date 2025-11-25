package nodes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dop251/goja"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type CodeJsNode struct{}

type codeJsData struct {
	Code  string                 `json:"code"`
	Input map[string]interface{} `json:"input"`
}

func (n *CodeJsNode) Execute(ctx context.Context, rawData []byte) (domain.NodeResult, error) {
	var data codeJsData
	if err := json.Unmarshal(rawData, &data); err != nil {
		return domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to parse input: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}

	vm := goja.New()

	// Set input variable
	vm.Set("input", data.Input)

	// Set console.log to capture logs
	var logs []string
	vm.Set("console", map[string]interface{}{
		"log": func(call goja.FunctionCall) goja.Value {
			var args []interface{}
			for _, arg := range call.Arguments {
				args = append(args, arg.Export())
			}
			logs = append(logs, fmt.Sprint(args...))
			return goja.Undefined()
		},
	})

	// Run the code
	val, err := vm.RunString(data.Code)
	if err != nil {
		return domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("JS Execution Error: %v\nLogs: %v", err, logs),
			OutputData:      map[string]interface{}{"error": err.Error()},
		}, nil
	}

	output := val.Export()

	// If output is not a map, wrap it
	var outputMap map[string]interface{}
	if m, ok := output.(map[string]interface{}); ok {
		outputMap = m
	} else {
		outputMap = map[string]interface{}{"result": output}
	}

	return domain.NodeResult{
		Status:          "completed",
		TriggeredHandle: "output_success",
		Log:             fmt.Sprintf("JS Execution Success. Logs: %v", logs),
		OutputData:      outputMap,
	}, nil
}
