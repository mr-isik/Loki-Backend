package engine

import (
	"github.com/mr-isik/loki-backend/internal/domain"
	"github.com/mr-isik/loki-backend/internal/engine/nodes"
)

// defaultRegistry is the package-level registry used by NewNodeExecutor.
var defaultRegistry = NewNodeRegistry()

func init() {
	// ── Trigger nodes ──────────────────────────────────────────────
	defaultRegistry.Register("webhook", func() domain.INodeExecutor { return &nodes.WebhookNode{} })
	defaultRegistry.Register("cron", func() domain.INodeExecutor { return &nodes.CronNode{} })

	// ── Action nodes ───────────────────────────────────────────────
	defaultRegistry.Register("http_request", func() domain.INodeExecutor { return &nodes.HttpRequestNode{} })
	defaultRegistry.Register("shell_command", func() domain.INodeExecutor { return &nodes.ShellCommandNode{} })
	defaultRegistry.Register("code_js", func() domain.INodeExecutor { return &nodes.CodeJsNode{} })
	defaultRegistry.Register("email_smtp", func() domain.INodeExecutor { return &nodes.EmailSmtpNode{} })
	defaultRegistry.Register("slack", func() domain.INodeExecutor { return &nodes.SlackNode{} })

	// ── Control-flow nodes ─────────────────────────────────────────
	defaultRegistry.Register("condition", func() domain.INodeExecutor { return &nodes.ConditionNode{} })
	defaultRegistry.Register("loop", func() domain.INodeExecutor { return &nodes.LoopNode{} })
	defaultRegistry.Register("wait", func() domain.INodeExecutor { return &nodes.WaitNode{} })
	defaultRegistry.Register("merge", func() domain.INodeExecutor { return &nodes.MergeNode{} })

	// ── Data / utility nodes ───────────────────────────────────────
	defaultRegistry.Register("set_data", func() domain.INodeExecutor { return &nodes.SetDataNode{} })
	defaultRegistry.Register("log", func() domain.INodeExecutor { return &nodes.LogNode{} })

	// ── File nodes ─────────────────────────────────────────────────
	defaultRegistry.Register("file_read", func() domain.INodeExecutor { return &nodes.FileReadNode{} })
	defaultRegistry.Register("file_write", func() domain.INodeExecutor { return &nodes.FileWriteNode{} })

	// ── Database nodes ─────────────────────────────────────────────
	defaultRegistry.Register("db_postgres", func() domain.INodeExecutor { return &nodes.DbPostgresNode{} })
	defaultRegistry.Register("db_mysql", func() domain.INodeExecutor { return &nodes.DbMysqlNode{} })

	// ── Message-queue nodes ────────────────────────────────────────
	defaultRegistry.Register("mq_rabbitmq_publish", func() domain.INodeExecutor { return &nodes.MqRabbitmqPublishNode{} })
}

// NewNodeExecutor returns a new INodeExecutor for the given type key.
// It delegates to the default package-level registry.
func NewNodeExecutor(typeKey string) (domain.INodeExecutor, error) {
	return defaultRegistry.Get(typeKey)
}

// RegisterNode allows external packages to register custom node executors
// at runtime without modifying this file (Open-Closed Principle).
func RegisterNode(typeKey string, factory NodeFactory) {
	defaultRegistry.Register(typeKey, factory)
}

// DefaultRegistry exposes the package-level registry for advanced use cases
// such as listing all registered types or checking availability.
func DefaultRegistry() *NodeRegistry {
	return defaultRegistry
}
