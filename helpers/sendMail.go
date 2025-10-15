package helpers

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendEmail(to string, otp int) error {
	from := os.Getenv("MAIL_SEND")
	subject := "Mã otp xác thực đổi mật khẩu"
	body := fmt.Sprintf("Mã otp của bạn là: %d", otp)

	msg := []byte(fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\n\r\n%s", to, from, subject, body))

	auth := smtp.PlainAuth("", os.Getenv("MAIL_SEND"), os.Getenv("MAIL_SECRET"), "smtp.gmail.com")
	return smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, msg)
}
