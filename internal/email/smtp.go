package email

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/eachchat/mailbot/internal/db"
	"github.com/eachchat/mailbot/pkg/utils"
	"maunium.net/go/mautrix/id"
)

func SetupSmtp(msg []string, roomID id.RoomID) error {
	var ign int
	var port int
	var host string
	if len(msg) < 4 {
		return fmt.Errorf("wrong command: !setup smtp,smtp.example.com:587,user@exapmole.com,PASSWORD,ture/false")
	}
	if isSMTPAccountAlreadyInUse(msg[1]) {
		return fmt.Errorf("smtp has already set up, the mail: %s", msg[1])
	}
	hostPort := strings.Split(msg[0], ":")
	if len(hostPort) < 2 {
		return fmt.Errorf("wrong host: !setup smtp,smtp.example.com:587,user@exapmole.com,PASSWORD,ture/false")
	} else {
		host = hostPort[0]
		tport, err := strconv.Atoi(hostPort[1])
		port = tport
		if err != nil {
			return fmt.Errorf("wrong port format: !setup smtp,smtp.example.com:587,user@exapmole.com,PASSWORD,ture/false")
		}
	}
	if strings.TrimSpace(strings.ToLower(msg[3])) == "true" {
		ign = 1
	}
	insertSMTPAccountount(roomID.String(), host, port, msg[1], utils.B64Encode(msg[2]), ign)

	return nil
}
func insertSMTPAccountount(roomID, host string, port int, username, password string, ignoreSSL int) (err error) {
	tx := DB.Create(&db.SmtpAccounts{RoomID: roomID, Host: host, Port: port, UserName: username, Password: password, IgnoreSSL: ignoreSSL})
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}
func isSMTPAccountAlreadyInUse(email string) bool {
	var count int64
	DB.Model(&db.SmtpAccounts{}).Where("user_name = ?", email).Count(&count)
	return count >= 1
}

func RemoveSMTPAccount(roomID string) error {
	tx := DB.Model(&db.SmtpAccounts{}).Where("room_id = ?", roomID).Delete(&db.SmtpAccounts{})
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func GetSMTPAccount(roomID string) (*db.SmtpAccounts, error) {
	var smtpA db.SmtpAccounts
	tx := DB.Model(smtpA).Where("room_id = ?", roomID).First(&smtpA)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &smtpA, nil
}
