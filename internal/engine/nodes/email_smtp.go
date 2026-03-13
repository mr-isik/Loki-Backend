package nodes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/mr-isik/loki-backend/internal/domain"
)

type EmailSmtpNode struct{}

type Attachment struct {
	Filename      string `json:"filename"`
	ContentBase64 string `json:"content_base64"`
}

type emailData struct {
	Host        string       `json:"host"`
	Port        int          `json:"port"`
	Username    string       `json:"username"`
	Password    string       `json:"password"`
	From        string       `json:"from"`
	To          []string     `json:"to"`
	Subject     string       `json:"subject"`
	Body        string       `json:"body"`
	HtmlBody    string       `json:"html_body"`
	Attachments []Attachment `json:"attachments"`
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

	var msg bytes.Buffer
	boundary := "MixedBoundaryString"
	altBoundary := "AlternativeBoundaryString"

	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(data.To, ",")))
	msg.WriteString(fmt.Sprintf("From: %s\r\n", data.From))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", data.Subject))
	msg.WriteString("MIME-Version: 1.0\r\n")

	if len(data.Attachments) > 0 {
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", boundary))
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	}

	if data.HtmlBody != "" && data.Body != "" {
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", altBoundary))
		msg.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
		msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n")
		msg.WriteString(data.Body + "\r\n")
		msg.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
		msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n\r\n")
		msg.WriteString(data.HtmlBody + "\r\n")
		msg.WriteString(fmt.Sprintf("--%s--\r\n", altBoundary))
	} else if data.HtmlBody != "" {
		msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n\r\n")
		msg.WriteString(data.HtmlBody + "\r\n")
	} else {
		msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n")
		msg.WriteString(data.Body + "\r\n")
	}

	if len(data.Attachments) > 0 {
		for _, att := range data.Attachments {
			msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			msg.WriteString(fmt.Sprintf("Content-Type: application/octet-stream; name=\"%s\"\r\n", att.Filename))
			msg.WriteString("Content-Transfer-Encoding: base64\r\n")
			msg.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", att.Filename))
			msg.WriteString(att.ContentBase64 + "\r\n")
		}
		msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	}

	err := smtp.SendMail(addr, auth, data.From, data.To, msg.Bytes())
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
