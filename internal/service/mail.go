package service

func NewMailService() *MailService {
	return &MailService{}
}

type MailService struct {
}

func (m *MailService) SendNoReplyEmail(emailAddress, subject, content string) error {
	return nil
}
