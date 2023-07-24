package email

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/eachchat/mailbot/internal/db"
	"github.com/eachchat/mailbot/pkg/utils"
	"gopkg.in/gomail.v2"
)

func SendMailMsg(msg string, roomID string, account *db.SmtpAccounts) error {
	m := gomail.NewMessage()
	m.SetHeader("From", account.UserName)
	msg = strings.Replace(msg, "!send", "", 1)
	msgSlice := strings.Split(strings.TrimSpace(msg), "\n")
	to := msgSlice[0]
	if !checkMailAddress(to) {
		return fmt.Errorf("wrong email address: %v", to)
	}
	m.SetHeader("To", strings.Split(to, ",")...)
	for i, v := range msgSlice[1:] {
		vSlice := strings.Split(strings.TrimSpace(v), " ")
		switch vSlice[0] {
		case "Cc":
			mList := strings.Split(strings.Join(vSlice[1:], " "), ",")
			m.SetHeader("Cc", mList...)
		case "CC":
			mList := strings.Split(strings.Join(vSlice[1:], " "), ",")
			m.SetHeader("Cc", mList...)
		case "Bcc":
			mList := strings.Split(strings.Join(vSlice[1:], " "), ",")
			m.SetHeader("Bcc", mList...)
		case "BCC":
			mList := strings.Split(strings.Join(vSlice[1:], " "), ",")
			m.SetHeader("Bcc", mList...)
		case "Subject":
			m.SetHeader("Subject", strings.Join(vSlice[1:], " "))
		default:
			m.SetBody("text/plain", strings.Join(msgSlice[i-1:], "\n"))
		}
	}
	LOG.Debug().Msg(fmt.Sprintf("Host: %v, Port: %v, User: %v", account.Host, account.Port, account.UserName))
	gd := gomail.NewDialer(account.Host, account.Port, account.UserName, utils.B64Decode(account.Password))
	if account.IgnoreSSL == 1 {
		gd.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if err := gd.DialAndSend(m); err != nil {
		fmt.Println(err.Error())
		LOG.Error().Msg(fmt.Sprintf("Failed to Send mail, error: %s", err.Error()))
		return fmt.Errorf("failed to send email, please check you smtp account, use !logout to exit current smtp and then !setup again")
	}
	return nil
}

func checkMailAddress(ma string) bool {
	for _, v := range strings.Split(ma, ",") {
		if !strings.Contains(v, "@") {
			return false
		}
	}
	return true
}
