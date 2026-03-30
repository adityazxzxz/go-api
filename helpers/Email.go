package helpers

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/gomail.v2"
)

func SendEmail(to string, subject string, body string) error {

	from := os.Getenv("EMAIL_SENDER")
	password := os.Getenv("EMAIL_PASSWORD")
	host := os.Getenv("EMAIL_HOST")
	portString := os.Getenv("EMAIL_PORT")
	port, err := strconv.Atoi(portString)
	if err != nil {
		fmt.Println("ERROR: Invalid email port")
		return err
	}

	// buat message
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// dialer (SESUAI OUTLOOK)
	d := gomail.NewDialer(
		host,
		port,
		from,
		password,
	)

	// kirim email
	if err := d.DialAndSend(m); err != nil {
		fmt.Println("ERROR:", err)
		return nil
	}
	return nil
}

func MailTemplateFormat(data map[string]interface{}, template string) string {
	result := template

	for key, value := range data {
		placeholder := "$" + key + "$"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}

	return result
}
