package notifications

import (
	"fmt"

	dbops "github.com/jacewalker/ip-monitor/db"
	"github.com/smtp2go-oss/smtp2go-go"
)

func SendEmailNotification(ch dbops.Check) bool {

	to := fmt.Sprintf(ch.Email)
	subject := fmt.Sprintf("Open Ports have changed for %s", ch.Address)
	textbody := fmt.Sprintf("Open Ports have changed for %s", ch.Address)
	htmlbody := fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
                <head>
                        <meta charset="utf-8">
                        <title>Open Ports have changed for %s</title>
                </head>
                <body style="font-family: sans-serif;">
                        <h1>Open Ports have changed for %s</h1>
                        <p>Below are the new open ports:</p>
                        <p>%s</p>
                        <p>Kind regards,</p>
                        <p>Jace's IP Monitoring</p>
                </body>
        </html>

`, ch.Address, ch.Address, ch.OpenPorts)

	email := smtp2go.Email{
		From: "IP Monitoring <ipmon@jcwlkr.io>",
		To: []string{
			to,
		},
		Subject:  subject,
		TextBody: textbody,
		HtmlBody: htmlbody,
	}
	result, err := smtp2go.Send(&email)
	if err != nil || result.Data.Error != "" {
		fmt.Println("An Error Occurred:", err)
		fmt.Println("SMTP2Go Data Error: " + result.Data.Error)
		fmt.Println("SMTP2Go Data Error Code: " + result.Data.ErrorCode)
		fmt.Println("SMTP2Go Field Validation Error - Field Name: " + result.Data.FieldValidationErrors.FieldName)
		fmt.Println("SMTP2Go Field Validation Error - Message: " + result.Data.FieldValidationErrors.Message)
		return false
	}
	fmt.Println("Email sent successfully.")
	return true
}
