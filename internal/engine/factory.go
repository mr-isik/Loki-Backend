package engine

import (
	"fmt"

	"github.com/mr-isik/loki-backend/internal/domain"
	"github.com/mr-isik/loki-backend/internal/engine/nodes"
)

func NewNodeExecutor(typeKey string) (domain.INodeExecutor, error) {
	switch typeKey {
	case "http_request":
		return &nodes.HttpRequestNode{}, nil
	case "shell_command":
		return &nodes.ShellCommandNode{}, nil
	case "condition":
		return &nodes.ConditionNode{}, nil
	case "loop":
		return &nodes.LoopNode{}, nil
	case "webhook":
		return &nodes.WebhookNode{}, nil
	case "cron":
		return &nodes.CronNode{}, nil
	case "wait":
		return &nodes.WaitNode{}, nil
	case "merge":
		return &nodes.MergeNode{}, nil
	case "set_data":
		return &nodes.SetDataNode{}, nil
	case "code_js":
		return &nodes.CodeJsNode{}, nil
	case "log":
		return &nodes.LogNode{}, nil
	case "file_read":
		return &nodes.FileReadNode{}, nil
	case "file_write":
		return &nodes.FileWriteNode{}, nil
	case "db_postgres":
		return &nodes.DbPostgresNode{}, nil
	case "db_mysql":
		return &nodes.DbMysqlNode{}, nil
	case "email_smtp":
		return &nodes.EmailSmtpNode{}, nil
	case "slack":
		return &nodes.SlackNode{}, nil
	case "mq_rabbitmq_publish":
		return &nodes.MqRabbitmqPublishNode{}, nil
	default:
		return nil, fmt.Errorf("unknown node type: %s", typeKey)
	}
}
