package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

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
	return m.sendMailMailGun("no-reply", recipientName, emailAddress, subject, content)
}

func (m *MailService) sendMailMailGun(from, recipientName, emailAddress, subject, content string) error {
	endpoint := "https://api.eu.mailgun.net/v3/toolsofworship.com/messages"

	data := url.Values{}
	data.Set("from", "Tools of Worship - no-reply <"+from+"@"+m.serverConfig.GetDomain()+">")
	data.Set("to", recipientName+"<"+emailAddress+">")
	data.Set("subject", subject)
	data.Set("html", content)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return fmt.Errorf("Error creating request: %v", err)
	}

	// Basic auth: username "api", password is your API key
	req.SetBasicAuth("api", m.config.GetMailKey())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Request failed:", err)
		return fmt.Errorf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Status:", resp.Status)
	fmt.Println("Body:", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Mailgun API returned status: %s", resp.Status)
	}

	fmt.Println("Email sent successfully.")

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
			"name":  "Tools of Worship - no-reply",
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

	resp, err := http.DefaultClient.Do(req)
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
