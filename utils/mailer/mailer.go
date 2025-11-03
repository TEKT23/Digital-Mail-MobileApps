package mailer

import (
	"bytes"
	"crypto/tls"
	"embed"
	"fmt"
	"html/template"
	"net/smtp"

	"TugasAkhir/config"
)

var (
	//go:embed templates/password_reset.html
	emailTemplates embed.FS

	passwordResetTemplate = template.Must(template.New("password_reset.html").ParseFS(emailTemplates, "templates/password_reset.html"))
)

type Client struct {
	cfg config.EmailConfig
}

func NewClient(cfg config.EmailConfig) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) SendPasswordResetEmail(toEmail, resetLink string) error {
	if c.cfg.Host == "" {
		return fmt.Errorf("smtp host is not configured")
	}

	from := c.cfg.FromAddress
	if from == "" {
		from = c.cfg.Username
	}
	if from == "" {
		return fmt.Errorf("smtp from address is not configured")
	}

	addr := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)
	auth := smtp.PlainAuth("", c.cfg.Username, c.cfg.Password, c.cfg.Host)

	body := bytes.Buffer{}
	data := struct {
		ResetLink string
	}{ResetLink: resetLink}

	if err := passwordResetTemplate.Execute(&body, data); err != nil {
		return fmt.Errorf("render password reset template: %w", err)
	}

	msg := buildHTMLMessage(from, toEmail, "Reset Password", body.String())

	if c.cfg.Username == "" && c.cfg.Password == "" {
		return smtp.SendMail(addr, nil, from, []string{toEmail}, []byte(msg))
	}

	if c.cfg.Port == 465 {
		return c.sendSMTPTLS(addr, auth, from, toEmail, msg)
	}

	return smtp.SendMail(addr, auth, from, []string{toEmail}, []byte(msg))
}

func (c *Client) sendSMTPTLS(addr string, auth smtp.Auth, from, toEmail, msg string) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: c.cfg.Host})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, c.cfg.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(toEmail); err != nil {
		return err
	}

	wc, err := client.Data()
	if err != nil {
		return err
	}
	_, err = wc.Write([]byte(msg))
	if err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}

	return client.Quit()
}
func buildHTMLMessage(from, to, subject, htmlBody string) string {
	return fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"utf-8\"\r\n\r\n%s", from, to, subject, htmlBody)
}
