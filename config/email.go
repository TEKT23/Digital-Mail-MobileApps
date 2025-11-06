package config

import (
	"os"
	"strconv"
)

type EmailConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
}

func LoadEmailConfig() EmailConfig {

	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil || port <= 0 {
		port = 587
	}

	return EmailConfig{
		Host:        os.Getenv("SMTP_HOST"),
		Port:        port,
		Username:    os.Getenv("SMTP_USERNAME"),
		Password:    os.Getenv("SMTP_PASSWORD"),
		FromAddress: os.Getenv("SMTP_FROM"),
	}
}
