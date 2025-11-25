package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/mr-isik/loki-backend/internal/domain"
)

type EmailSmtpNode struct{}

type emailData struct {
	Host     string   `json:"host"`
	Port     int      `json:"port"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	From     string   `json:"from"`
	To       []string `json:"to"`
	Subject  string   `json:"subject"`
	Body     string   `json:"body"`
}

func (n *EmailSmtpNode) Execute(ctx context.Context, rawData []byte) (*domain.NodeResult, error) {
	var data emailData
	if err := json.Unmarshal(rawData, &data); err != nil {
		return &domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to parse input: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}

	auth := smtp.PlainAuth("", data.Username, data.Password, data.Host)
	addr := fmt.Sprintf("%s:%d", data.Host, data.Port)

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", strings.Join(data.To, ","), data.Subject, data.Body))

	err := smtp.SendMail(addr, auth, data.From, data.To, msg)
	if err != nil {
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Failed to send email: %v", err),
			OutputData:      map[string]interface{}{"error": err.Error()},
		}, nil
	}

	return &domain.NodeResult{
		Status:          "completed",
		TriggeredHandle: "output_success",
		Log:             fmt.Sprintf("Email sent to %v", data.To),
		OutputData:      map[string]interface{}{"sent": true},
	}, nil
}
