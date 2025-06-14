package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/config"
)

func NewMailService(config config.MailConfig, serverConfig config.ServerConfig) *MailService {
	return &MailService{config: config, serverConfig: serverConfig}
}

type MailService struct {
	config       config.MailConfig
	serverConfig config.ServerConfig
}

func (m *MailService) SendNoReplyEmail(recipientName, emailAddress, subject, content string) error {
	url := "https://api.sendgrid.com/v3/mail/send"

	payload := map[string]interface{}{
		"personalizations": []map[string]interface{}{
			{
				"to": []map[string]string{
					{"email": emailAddress, "name": recipientName},
				},
				"subject": subject,
			},
		},
		"from": map[string]string{
			"email": "no-reply@" + m.serverConfig.GetDomain(),
			"name":  "Tools of Worship",
		},
		"content": []map[string]string{
			{
				"type":  "text/html",
				"value": content,
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return fmt.Errorf("Error marshalling JSON: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+m.config.GetMailKey())
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Request failed:", err)
		return fmt.Errorf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Status:", resp.Status)
	fmt.Println("Body:", string(body))

	return nil
}
