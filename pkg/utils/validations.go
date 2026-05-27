package utils

import (
	"fmt"
	emailverifier "github.com/AfterShip/email-verifier"
	"net"
	"net/mail"
	"strings"
)

var (
	verifier = emailverifier.NewVerifier().EnableSMTPCheck()
)

func VerifyEmail(email string) error {
	_, err := verifier.Verify(email)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	_, err = mail.ParseAddress(email)
	if err != nil {
		return err
	}

	// Проверяем наличие MX-записей у домена
	domain := email[strings.LastIndex(email, "@")+1:]
	mxRecords, err := net.LookupMX(domain)
	if err != nil && len(mxRecords) < 0 {
		return err
	}

	return nil
}
