package mailer

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"TugasAkhir/config"
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

	addr := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)
	auth := smtp.PlainAuth("", c.cfg.Username, c.cfg.Password, c.cfg.Host)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: Reset Password\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"utf-8\"\r\n\r\nHalo,\r\n\r\nKami menerima permintaan untuk mengatur ulang kata sandi Anda. Silakan buka tautan berikut untuk melanjutkan proses reset kata sandi:\r\n%s\r\n\r\nJika Anda tidak meminta reset kata sandi, abaikan email ini.\r\n", c.cfg.FromAddress, toEmail, resetLink)

	if c.cfg.Username == "" && c.cfg.Password == "" {
		return smtp.SendMail(addr, nil, c.cfg.FromAddress, []string{toEmail}, []byte(msg))
	}

	if c.cfg.Port == 465 {
		return c.sendSMTPTLS(addr, auth, toEmail, msg)
	}

	return smtp.SendMail(addr, auth, c.cfg.FromAddress, []string{toEmail}, []byte(msg))
}

func (c *Client) sendSMTPTLS(addr string, auth smtp.Auth, toEmail, msg string) error {
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

	if err := client.Mail(c.cfg.FromAddress); err != nil {
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
