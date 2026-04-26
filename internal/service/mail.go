package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/config"
)

func NewMailService(config config.MailConfig, serverConfig config.ServerConfig, logger *log.Logger) *MailService {
	return &MailService{config: config, serverConfig: serverConfig, logger: logger}
}

type MailService struct {
	config       config.MailConfig
	serverConfig config.ServerConfig
	logger       *log.Logger
}

func (m *MailService) SendNoReplyEmail(recipientName, emailAddress, subject, content string) error {
	return m.sendMailMailGun("no-reply", recipientName, emailAddress, subject, content)
}

func (m *MailService) sendMailMailGun(from, recipientName, emailAddress, subject, content string) error {
	endpoint := m.config.GetMailEndpoint()

	data := url.Values{}
	data.Set("from", "Tools of Worship <"+from+"@"+m.config.GetMailDomain()+">")
	data.Set("to", recipientName+"<"+emailAddress+">")
	data.Set("subject", subject)
	data.Set("html", content)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		m.logger.Printf("mail: error creating request: %v", err)
		return fmt.Errorf("error creating request: %w", err)
	}

	// Basic auth: username "api", password is your API key
	req.SetBasicAuth("api", m.config.GetMailKey())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.logger.Printf("mail: request failed: %v", err)
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	m.logger.Printf("mail: %s body: %s", resp.Status, string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mailgun API returned status: %s", resp.Status)
	}

	m.logger.Println("mail: email sent successfully")

	return nil
}

func (m *MailService) sendMailSendGrid(from, recipientName, emailAddress, subject, content string) error {
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
			"email": from + "@" + m.serverConfig.GetDomain(),
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
		m.logger.Printf("mail: error marshalling JSON: %v", err)
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		m.logger.Printf("mail: error creating request: %v", err)
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+m.config.GetMailKey())
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.logger.Printf("mail: request failed: %v", err)
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	m.logger.Printf("mail: %s body: %s", resp.Status, string(body))

	return nil
}
